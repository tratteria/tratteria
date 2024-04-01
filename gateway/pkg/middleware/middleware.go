package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func Authenticate(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			if err != nil {
				logger.Error("Failed to retrieve session_token cookie", zap.Error(err))
			} else {
				logger.Error("session_token cookie is not present")
			}

			http.Error(w, "Unauthorized: Missing or invalid authentication cookie.", http.StatusUnauthorized)
			
			return
		}

        r.Header.Set("x-user-name", cookie.Value)
        next.ServeHTTP(w, r)
	})
}
