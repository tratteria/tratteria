package middleware

import (
	"net/http"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/trats"
)

func CombineMiddleware(middleware ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			final = middleware[i](final)
		}

		return final
	}
}

func GetMiddleware(stocksConfig *config.StocksConfig, spireJwtSource *workloadapi.JWTSource, traTsVerifier *trats.Verifier, logger *zap.Logger) func(http.Handler) http.Handler {
	middlewareList := []func(http.Handler) http.Handler{}

	if stocksConfig.Toggles.SpireToggle {
		middlewareList = append(middlewareList, getSpiffeMiddleware(stocksConfig, spireJwtSource, logger))
	}

	if stocksConfig.Toggles.TxnTokenToggle {
		middlewareList = append(middlewareList, getTxnTokenMiddleware(traTsVerifier, stocksConfig.TxnTokenKeys.JWKS, logger))
	}

	return CombineMiddleware(middlewareList...)
}
