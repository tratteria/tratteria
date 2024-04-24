package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/common"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"

	"go.uber.org/zap"
)

func getAuthenticationMiddleware(oauth2Config oauth2.Config, oidcProvider *oidc.Provider, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session_id, err := r.Cookie("session_id")
			if err != nil || session_id.Value == "" {
				if err != nil {
					logger.Error("Failed to retrieve session_id cookie", zap.Error(err))
				} else {
					logger.Error("session_id cookie is not present")
				}

				http.Error(w, "Unauthorized: Missing or invalid authentication cookie.", http.StatusUnauthorized)

				return
			}

			oidcConfig := &oidc.Config{
				ClientID: oauth2Config.ClientID,
			}
			verifier := oidcProvider.Verifier(oidcConfig)

			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)

			defer cancel()

			idToken, err := verifier.Verify(ctx, session_id.Value)
			if err != nil {
				logger.Error("Failed to verify session_id token.", zap.Error(err))
				http.Error(w, "Failed to verify session_id token", http.StatusUnauthorized)

				return
			}

			var claims common.IDTokenClaims

			if err := idToken.Claims(&claims); err != nil {
				logger.Error("Failed to parse session_id token claims.", zap.Error(err))
				http.Error(w, "Failed to parse session_id token claims", http.StatusUnauthorized)

				return
			}

			logger.Info("Session_id token verified successfully.", zap.String("email", claims.Email))

			ctx = context.WithValue(r.Context(), common.OIDC_ID_TOKEN_CONTEXT_KEY, session_id.Value)

			r = r.WithContext(ctx)

			r.Header.Set("x-user-name", claims.Email)

			next.ServeHTTP(w, r)
		})
	}
}
