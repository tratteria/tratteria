package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/tratteria/tratteria/pkg/accessevaluation"
	"github.com/tratteria/tratteria/pkg/common"
	"github.com/tratteria/tratteria/pkg/config"
	"github.com/tratteria/tratteria/pkg/keys"
	"github.com/tratteria/tratteria/pkg/subjecttokenhandler"
	"github.com/tratteria/tratteria/pkg/txntokenerrors"
	"go.uber.org/zap"
)

type Service struct {
	Config               *config.AppConfig
	SpireJwtSource       *workloadapi.JWTSource
	SubjectTokenHandlers *subjecttokenhandler.TokenHandlers
	AccessEvaluator      accessevaluation.AccessEvaluatorService
	Logger               *zap.Logger
}

func NewService(config *config.AppConfig, spireJwtSource *workloadapi.JWTSource, subjectTokenHandlers *subjecttokenhandler.TokenHandlers, accessEvaluator accessevaluation.AccessEvaluatorService, logger *zap.Logger) *Service {
	return &Service{
		Config:               config,
		SpireJwtSource:       spireJwtSource,
		SubjectTokenHandlers: subjectTokenHandlers,
		AccessEvaluator:      accessEvaluator,
		Logger:               logger,
	}
}

const (
	TOKEN_JWT_HEADER = "txn_token"
)

type TokenRequest struct {
	RequestedTokenType common.TokenType
	SubjectToken       string
	SubjectTokenType   common.TokenType
	Scope              string
	RequestDetails     map[string]any
	RequestContext     map[string]any
}

type TokenResponse struct {
	TokenType       string           `json:"token_type"`
	IssuedTokenType common.TokenType `json:"issued_token_type"`
	AccessToken     string           `json:"access_token"`
}

func (s *Service) GetJwks() jwk.Set {
	return keys.GetJWKS()
}

func (s *Service) GenerateTxnToken(ctx context.Context, txnTokenRequest *TokenRequest) (*TokenResponse, error) {
	subjectTokenHandler, err := s.SubjectTokenHandlers.GetHandler(txnTokenRequest.SubjectTokenType)
	if err != nil {
		s.Logger.Error("Failed to get subject token handler.", zap.String("subject-token-type", string(txnTokenRequest.SubjectTokenType)), zap.Error(err))

		return &TokenResponse{}, err
	}

	claims, err := subjectTokenHandler.VerifyAndParse(ctx, txnTokenRequest.SubjectToken)
	if err != nil {
		s.Logger.Error("Failed to verify and parse subject token.", zap.Error(err))

		return &TokenResponse{}, err
	}

	subject, err := subjectTokenHandler.ExtractSubject(claims)
	if err != nil {
		s.Logger.Error("Failed to extract subject.", zap.Error(err))

		return &TokenResponse{}, err
	}

	s.Logger.Info("Successfully verified subject token.", zap.Any("subject", subject))

	accessEvaluation, err := s.AccessEvaluator.Evaluate(claims, txnTokenRequest.Scope, txnTokenRequest.RequestDetails, txnTokenRequest.RequestContext)
	if err != nil {
		s.Logger.Error("Error evaluating access.", zap.Error(err))

		return &TokenResponse{}, err
	}

	if !accessEvaluation {
		s.Logger.Error("Access Denied.",
			zap.Any("subject", subject),
			zap.Any("scope", txnTokenRequest.Scope),
			zap.Any("request-details", txnTokenRequest.RequestDetails),
			zap.Any("request-context", txnTokenRequest.RequestContext),
		)

		return &TokenResponse{}, txntokenerrors.ErrAccessDenied
	}

	s.Logger.Info("Access authorized for request.", zap.Any("subject", subject), zap.String("scope", txnTokenRequest.Scope))

	txnID, err := uuid.NewRandom()
	if err != nil {
		s.Logger.Error("Error generating transaction id.")

		return &TokenResponse{}, err
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":  s.Config.Issuer,
		"iat":  time.Now().Unix(),
		"aud":  s.Config.Audience,
		"exp":  time.Now().Add(s.Config.Token.LifeTime).Unix(),
		"txn":  txnID,
		"sub":  subject,
		"purp": txnTokenRequest.Scope,
		"azd":  txnTokenRequest.RequestDetails,
		"rctx": txnTokenRequest.RequestContext,
	})

	newToken.Header["typ"] = TOKEN_JWT_HEADER
	newToken.Header["kid"] = keys.GetKid()

	privateKey := keys.GetPrivateKey()

	tokenString, err := newToken.SignedString(privateKey)
	if err != nil {
		s.Logger.Error("Failed to sign txn token.", zap.Error(err))

		return &TokenResponse{}, err
	}

	tokenResponse := &TokenResponse{
		TokenType:       "N_A",
		IssuedTokenType: common.TXN_TOKEN_TYPE,
		AccessToken:     tokenString,
	}

	return tokenResponse, nil
}
