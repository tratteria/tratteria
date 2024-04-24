package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"go.uber.org/zap"
)

func getJwtSvidMiddleware(targetServiceSpiffeID spiffeid.ID, spireJwtSource *workloadapi.JWTSource, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
}
