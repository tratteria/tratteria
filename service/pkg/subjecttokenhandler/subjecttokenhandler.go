package subjecttokenhandler

import (
	"context"
	"errors"
	"fmt"

	"github.com/tratteria/tratteria/pkg/common"
	"github.com/tratteria/tratteria/pkg/subjectidentifier"
	"go.uber.org/zap"
)

type SubjectTokens struct {
	OIDC       *OIDCToken       `json:"OIDC,omitempty"`
	SelfSigned *SelfSignedToken `json:"selfSigned,omitempty"`
}

type OIDCToken struct {
	ClientID     string `json:"clientId"`
	ProviderURL  string `json:"providerURL"`
	SubjectField string `json:"subjectField"`
}

type SelfSignedToken struct {
	Validation    bool   `json:"validation"`
	JWKSSEndpoint string `json:"jwksEndpoint"`
}

type TokenHandler interface {
	VerifyAndParse(ctx context.Context, token string) (interface{}, error)
	ExtractSubject(claims interface{}) (subjectidentifier.Identifier, error)
}

type TokenHandlers struct {
	oIDCTokenHandler       TokenHandler
	selfSignedTokenHandler TokenHandler
}

func NewTokenHandlers(subjectTokens SubjectTokens, logger *zap.Logger) *TokenHandlers {
	handlers := &TokenHandlers{}

	if subjectTokens.OIDC != nil {
		handlers.oIDCTokenHandler = NewOIDCTokenHandler(subjectTokens.OIDC, logger)
	}

	if subjectTokens.SelfSigned != nil {
		handlers.selfSignedTokenHandler = NewSelfSignedTokenHandler(subjectTokens.SelfSigned, logger)
	}

	return handlers
}

func (t *TokenHandlers) GetHandler(tokenType common.TokenType) (TokenHandler, error) {
	switch tokenType {
	case common.OIDC_ID_TOKEN_TYPE:
		if t.oIDCTokenHandler != nil {
			return t.oIDCTokenHandler, nil
		}

		return nil, errors.New("configuration not provided for OIDC subject token")
	case common.SELF_SIGNED_TOKEN_TYPE:
		if t.selfSignedTokenHandler != nil {
			return t.selfSignedTokenHandler, nil
		}

		return nil, errors.New("configuration not provided for self-signed subject token")

	default:
		return nil, fmt.Errorf("unsupported token type: %s", tokenType)
	}
}
