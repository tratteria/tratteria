package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func Authenticate(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Validate user session to authenticate the request
		next.ServeHTTP(w, r)
	})
}
