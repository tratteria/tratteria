package subjecttokenhandler

import (
	"context"
	"errors"
	"fmt"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/common"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/subjectidentifier"
	"go.uber.org/zap"
)

type TokenHandler interface {
	VerifyAndParse(ctx context.Context, token string) (interface{}, error)
	ExtractSubject(claims interface{}) (subjectidentifier.Identifier, error)
}

type TokenHandlers struct {
	oIDCTokenHandler TokenHandler
}

func GetTokenHandlers(clientAuthenticationMethods *config.ClientAuthenticationMethods, logger *zap.Logger) *TokenHandlers {
	handlers := &TokenHandlers{}

	if clientAuthenticationMethods.OIDC != nil {
		handlers.oIDCTokenHandler = NewOIDCTokenHandler(clientAuthenticationMethods.OIDC, logger)
	}

	return handlers
}

func (t *TokenHandlers) GetHandler(tokenType common.TokenType) (TokenHandler, error) {
	switch tokenType {
	case common.OIDC_ID_TOKEN_TYPE:
		if t.oIDCTokenHandler != nil {
			return t.oIDCTokenHandler, nil
		}

		return nil, errors.New("client authentication configuration not provided for OIDC")
	default:
		return nil, fmt.Errorf("unsupported token type: %s", tokenType)
	}
}
