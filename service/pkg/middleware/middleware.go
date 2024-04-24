package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/authz"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
)

func spiffeMiddleware(config *config.AppConfig, spireJwtSource *workloadapi.JWTSource, logger *zap.Logger) func(http.Handler) http.Handler {
	publicEndpoints := authz.GetPublicEndpoints()
	policies := authz.GetSpiffeAccessControlPolicies(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			routePath, err := mux.CurrentRoute(r).GetPathTemplate()
			if err != nil {
				logger.Error("Error retrieving the route path template:", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			if slices.Contains(publicEndpoints, routePath) {
				next.ServeHTTP(w, r)

				return
			}

			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token == "" {
				logger.Error("JWT-SVID token not provided.")
				http.Error(w, "Unauthorized: JWT-SVID token not provided", http.StatusUnauthorized)

				return
			}

			svid, err := jwtsvid.ParseAndValidate(token, spireJwtSource, []string{config.Spiffe.ServiceID.String()})
			if err != nil {
				logger.Error("Failed to validate JWT-SVID token.", zap.Error(err))
				http.Error(w, "Unauthorized: Invalid JWT-SVID token", http.StatusUnauthorized)

				return
			}

			logger.Info("Successfully authenticated a request.", zap.String("spiffeID", svid.ID.String()))

			if !authz.IsSpiffeIDAuthorized(routePath, r.Method, svid.ID, policies) {
				logger.Error("Unauthorized access attempt.", zap.String("spiffeID", svid.ID.String()), zap.String("path", routePath), zap.String("method", r.Method))
				http.Error(w, "Forbidden: Access not permited to the resource", http.StatusForbidden)

				return
			}

			logger.Info("Successfully authorized a request.", zap.String("spiffeID", svid.ID.String()))

			next.ServeHTTP(w, r)
		})
	}
}

func CombineMiddleware(middleware ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			final = middleware[i](final)
		}

		return final
	}
}

func GetMiddleware(txnTokenConfig *config.AppConfig, spireJwtSource *workloadapi.JWTSource, logger *zap.Logger) func(http.Handler) http.Handler {
	middlewareList := []func(http.Handler) http.Handler{
		spiffeMiddleware(txnTokenConfig, spireJwtSource, logger),
	}

	return CombineMiddleware(middlewareList...)
}
