package middlewares

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func AuthorizeSpiffeID(authorizedIDs func() ([]spiffeid.ID, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				http.Error(w, "No TLS client authentication provided", http.StatusUnauthorized)

				return
			}

			spiffeID, err := getSpiffeIDFromCert(r.TLS.PeerCertificates[0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)

				return
			}

			authorizedIDs, err := authorizedIDs()
			if err != nil {
				http.Error(w, "Error authorizing request", http.StatusInternalServerError)

				return
			}

			authorizedIDStrings := make([]string, len(authorizedIDs))
			for i, id := range authorizedIDs {
				authorizedIDStrings[i] = id.String()
			}

			for _, id := range authorizedIDStrings {
				if spiffeID == id {
					next.ServeHTTP(w, r)

					return
				}
			}

			http.Error(w, "Unauthorized SPIFFE ID", http.StatusForbidden)
		})
	}
}

func getSpiffeIDFromCert(cert *x509.Certificate) (string, error) {
	uris := cert.URIs
	if len(uris) == 0 {
		return "", fmt.Errorf("no URIs found in certificate")
	}

	spiffeID := uris[0].String()
	if !strings.HasPrefix(spiffeID, "spiffe://") {
		return "", fmt.Errorf("invalid SPIFFE ID format")
	}

	return spiffeID, nil
}
