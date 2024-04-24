package subjecttokenhandler

import (
	"context"
	"fmt"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/subjectidentifier"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/txntokenerrors"
	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const OIDC_PROVIDER_INITILIZATION_MAX_RETRIES = 5

type OIDCTokenHandler struct {
	subjectField string
	verifier     *oidc.IDTokenVerifier
}

func NewOIDCTokenHandler(oidcConfig *config.OIDC, logger *zap.Logger) *OIDCTokenHandler {
	provider := getOIDCProvider(oidcConfig.ProviderURL, logger)

	verifier := provider.Verifier(&oidc.Config{
		ClientID: oidcConfig.ClientID,
	})

	return &OIDCTokenHandler{subjectField: oidcConfig.SubjectField,
		verifier: verifier}
}

func (o *OIDCTokenHandler) VerifyAndParse(ctx context.Context, token string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	idToken, err := o.verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	var claims jwt.MapClaims

	if err := idToken.Claims(&claims); err != nil {
		return nil, txntokenerrors.ErrInvalidSubjectTokenClaims
	}

	return claims, nil
}

func (o *OIDCTokenHandler) ExtractSubject(claims interface{}) (subjectidentifier.Identifier, error) {
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return nil, txntokenerrors.ErrInvalidSubjectTokenClaims
	}

	subjectValue, ok := mapClaims[o.subjectField].(string)
	if !ok {
		return nil, txntokenerrors.ErrConfiguredSubjectFieldNotFound
	}

	return subjectidentifier.NewIdentifier(o.subjectField, subjectValue)
}

func getOIDCProvider(oidcIssuer string, logger *zap.Logger) *oidc.Provider {
	delay := time.Second

	for i := 0; i < OIDC_PROVIDER_INITILIZATION_MAX_RETRIES; i++ {
		ctx := context.Background()

		provider, err := oidc.NewProvider(ctx, oidcIssuer)
		if err == nil {
			logger.Info("Successfully connected to the OIDC provider.")

			return provider
		}

		logger.Error("Failed to connect to the OIDC provider.",
			zap.Int("attempt", i+1),
			zap.String("retrying_in", delay.String()),
			zap.Error(err))
		time.Sleep(delay)

		delay *= 2
	}

	logger.Error(fmt.Sprintf("Failed to connect to the OIDC provider after %d attempts", OIDC_PROVIDER_INITILIZATION_MAX_RETRIES))

	panic(fmt.Sprintf("failed to connect to the OIDC provider after %d attempts", OIDC_PROVIDER_INITILIZATION_MAX_RETRIES))
}
