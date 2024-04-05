package middleware

import (
	"net/http"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/sessionmanager"
	"time"

	"go.uber.org/zap"
)

func Authenticate(next http.Handler, logger *zap.Logger) http.Handler {
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
            logger.Error("Session has expired", zap.String("session_id", cookie.Value))
            sessionmanager.DeleteSession(cookie.Value)
            http.Error(w, "Unauthorized: Session has expired.", http.StatusUnauthorized)
            
			return
        }

        r.Header.Set("x-user-name", sessionData.Email)
        
        next.ServeHTTP(w, r)
    })
}
