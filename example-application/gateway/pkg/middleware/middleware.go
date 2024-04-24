package middleware

import (
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"golang.org/x/oauth2"

	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/middleware/txntokenmiddleware"
)

func CombineMiddleware(middleware ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			final = middleware[i](final)
		}

		return final
	}
}

func GetMiddleware(oauth2Config oauth2.Config, oidcProvider *oidc.Provider, targetServiceSpiffeID spiffeid.ID, spireJwtSource *workloadapi.JWTSource, txnTokenServiceURL string, txnTokenServiceSpiffeID spiffeid.ID, httpClient *http.Client, logger *zap.Logger) func(http.Handler) http.Handler {
	middlewareList := []func(http.Handler) http.Handler{
		getAuthenticationMiddleware(oauth2Config, oidcProvider, logger),
		txntokenmiddleware.GetTxnTokenMiddleware(txnTokenServiceURL, httpClient, spireJwtSource, txnTokenServiceSpiffeID, logger),
		getJwtSvidMiddleware(targetServiceSpiffeID, spireJwtSource, logger),
	}

	return CombineMiddleware(middlewareList...)
}