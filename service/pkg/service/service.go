package service

import (
	"context"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/accessevaluation"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/common"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/generationrules/v1alpha1"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/keys"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/subjecttokenhandler"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/tratgenerator"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/txntokenerrors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
)

type Service struct {
	config                *config.AppConfig
	spireJwtSource        *workloadapi.JWTSource
	subjectTokenHandlers  *subjecttokenhandler.TokenHandlers
	generationRuleManager v1alpha1.GenerationRulesManager
	tratGenerator         *tratgenerator.TraTGenerator
	accessEvaluator       accessevaluation.AccessEvaluatorService
	logger                *zap.Logger
}

func NewService(config *config.AppConfig, spireJwtSource *workloadapi.JWTSource, subjectTokenHandlers *subjecttokenhandler.TokenHandlers, generationRuleManager v1alpha1.GenerationRulesManager, tratGenerator *tratgenerator.TraTGenerator, accessEvaluator accessevaluation.AccessEvaluatorService, logger *zap.Logger) *Service {
	return &Service{
		config:                config,
		spireJwtSource:        spireJwtSource,
		subjectTokenHandlers:  subjectTokenHandlers,
		generationRuleManager: generationRuleManager,
		tratGenerator:         tratGenerator,
		accessEvaluator:       accessEvaluator,
		logger:                logger,
	}
}

const (
	TOKEN_JWT_HEADER = "txn_token"
)

type TokenResponse struct {
	TokenType       string           `json:"token_type"`
	IssuedTokenType common.TokenType `json:"issued_token_type"`
	AccessToken     string           `json:"access_token"`
}

func (s *Service) GetJwks() jwk.Set {
	return keys.GetJWKS()
}

func (s *Service) GenerateTxnToken(ctx context.Context, txnTokenRequest *common.TokenRequest) (*TokenResponse, error) {
	subjectTokenHandler, err := s.subjectTokenHandlers.GetHandler(txnTokenRequest.SubjectTokenType)
	if err != nil {
		s.logger.Error("Failed to get subject token handler.", zap.String("subject-token-type", string(txnTokenRequest.SubjectTokenType)), zap.Error(err))

		return &TokenResponse{}, err
	}

	subjectTokenClaims, err := subjectTokenHandler.VerifyAndParse(ctx, txnTokenRequest.SubjectToken)
	if err != nil {
		s.logger.Error("Failed to verify and parse subject token.", zap.Error(err))

		return &TokenResponse{}, err
	}

	subject, err := subjectTokenHandler.ExtractSubject(subjectTokenClaims)
	if err != nil {
		s.logger.Error("Failed to extract subject.", zap.Error(err))

		return &TokenResponse{}, err
	}

	s.logger.Info("Successfully verified subject token.", zap.Any("subject", subject))

	scope, adz, err := s.tratGenerator.GenerateTraT(
		txnTokenRequest.RequestDetails.Path,
		txnTokenRequest.RequestDetails.Method,
		txnTokenRequest.RequestDetails.QueryParameters,
		txnTokenRequest.RequestDetails.Headers,
		txnTokenRequest.RequestDetails.Body,
	)
	if err != nil {
		s.logger.Error("Failed to generate scope and authorization details for a request.", zap.Error(err))

		return &TokenResponse{}, err
	}

	accessEvaluation, err := s.accessEvaluator.Evaluate(txnTokenRequest, subjectTokenClaims, scope, adz)
	if err != nil {
		s.logger.Error("Error evaluating access.", zap.Error(err))

		return &TokenResponse{}, err
	}

	if !accessEvaluation {
		s.logger.Error("Access Denied.",
			zap.Any("subject", subject),
			zap.Any("scope", scope),
		)

		return &TokenResponse{}, txntokenerrors.ErrAccessDenied
	}

	s.logger.Info("Access authorized for request.", zap.Any("subject", subject), zap.String("scope", scope))

	txnID, err := uuid.NewRandom()
	if err != nil {
		s.logger.Error("Error generating transaction id.")

		return &TokenResponse{}, err
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":  s.config.Issuer,
		"iat":  time.Now().Unix(),
		"aud":  s.config.Audience,
		"exp":  time.Now().Add(s.config.Token.LifeTime).Unix(),
		"txn":  txnID,
		"sub":  subject,
		"purp": scope,
		"azd":  adz,
		"rctx": txnTokenRequest.RequestContext,
	})

	newToken.Header["typ"] = TOKEN_JWT_HEADER
	newToken.Header["kid"] = keys.GetKid()

	privateKey := keys.GetPrivateKey()

	tokenString, err := newToken.SignedString(privateKey)
	if err != nil {
		s.logger.Error("Failed to sign txn token.", zap.Error(err))

		return &TokenResponse{}, err
	}

	tokenResponse := &TokenResponse{
		TokenType:       "N_A",
		IssuedTokenType: common.TXN_TOKEN_TYPE,
		AccessToken:     tokenString,
	}

	return tokenResponse, nil
}

func (s *Service) AddGenerationRule(pushedGenerationRule v1alpha1.GenerationRule) {
	s.generationRuleManager.AddRule(pushedGenerationRule)
}

func (s *Service) GetGenerationRules() map[string]map[string]v1alpha1.GenerationRule {
	return s.generationRuleManager.GetRules()
}
