package service

import (
	"context"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/common"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/keys"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/subjecttokenhandler"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
)

type Service struct {
	Config               *config.AppConfig
	SpireJwtSource       *workloadapi.JWTSource
	SubjectTokenHandlers *subjecttokenhandler.TokenHandlers
	Logger               *zap.Logger
}

func NewService(config *config.AppConfig, spireJwtSource *workloadapi.JWTSource, subjectTokenHandlers *subjecttokenhandler.TokenHandlers, logger *zap.Logger) *Service {
	return &Service{
		Config:               config,
		SpireJwtSource:       spireJwtSource,
		SubjectTokenHandlers: subjectTokenHandlers,
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
	IssuedTokenType common.TokenType `json:"issued_token_type"`
	AccessToken     string           `json:"access_token"`
}

func (s *Service) GetJwks() keys.JWKS {
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

	txnID, err := uuid.NewRandom()
	if err != nil {
		s.Logger.Error("Error generating transaction id.")

		return &TokenResponse{}, err
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":  s.Config.Issuer,
		"iat":  time.Now().Unix(),
		"aud":  s.Config.Audience,
		"exp":  time.Now().Add(time.Duration(time.Second)).Unix(),
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
		IssuedTokenType: common.TXN_TOKEN_TYPE,
		AccessToken:     tokenString,
	}

	return tokenResponse, nil
}