package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/sessionmanager"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"go.uber.org/zap"
)

func authenticate(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			if err != nil {
				logger.Error("Failed to retrieve session_id cookie", zap.Error(err))
			} else {
				logger.Error("session_id cookie is not present")
			}

			http.Error(w, "Unauthorized: Missing or invalid authentication cookie.", http.StatusUnauthorized)

			return
		}

		sessionData, exists := sessionmanager.GetSession(cookie.Value)
		if !exists {
			logger.Error("Session does not exist", zap.String("session_id", cookie.Value))
			http.Error(w, "Unauthorized: Session does not exist.", http.StatusUnauthorized)

			return
		}

		if sessionData.Expires.Before(time.Now()) {
            sessionmanager.DeleteSession(cookie.Value)
			logger.Error("Session has expired", zap.String("session_id", cookie.Value))
			http.Error(w, "Unauthorized: Session has expired.", http.StatusUnauthorized)

			return
		}

		r.Header.Set("x-user-name", sessionData.Email)

		next.ServeHTTP(w, r)
	})
}

func addJWTSVID(next http.Handler, targetServiceSpiffeID spiffeid.ID, spireJwtSource *workloadapi.JWTSource, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		svid, err := spireJwtSource.FetchJWTSVID(ctx, jwtsvid.Params{
			Audience: targetServiceSpiffeID.String(),
		})
		if err != nil {
			logger.Error("Failed to fetch JWT-SVID.", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		r.Header.Set("Authorization", "Bearer "+svid.Marshal())

		next.ServeHTTP(w, r)
	})
}

func GatewayMiddleware(next http.Handler, targetServiceSpiffeID spiffeid.ID, spireJwtSource *workloadapi.JWTSource, logger *zap.Logger) http.Handler {
	authenticateMiddleware := authenticate(next, logger)
	
	addJwtsvidMiddleware := addJWTSVID(authenticateMiddleware, targetServiceSpiffeID, spireJwtSource, logger)
	
	return addJwtsvidMiddleware
}