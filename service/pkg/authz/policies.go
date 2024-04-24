package authz

import (
	"net/http"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func GetPublicEndpoints() []string {
	return []string{
		"/.well-known/jwks.json",
	}
}

func GetSpiffeAccessControlPolicies(config *config.AppConfig) map[string]map[string][]spiffeid.ID {
	return map[string]map[string][]spiffeid.ID{
		"/token_endpoint": {http.MethodPost: config.Spiffe.AuthorizedServiceIDs},
	}
}

func IsSpiffeIDAuthorized(path, method string, spiffeID spiffeid.ID, policies map[string]map[string][]spiffeid.ID) bool {
	allowedMethods, ok := policies[path]
	if !ok {
		return false
	}

	allowedSpiffeIDs, ok := allowedMethods[method]
	if !ok {
		return false
	}

	for _, id := range allowedSpiffeIDs {
		if id == spiffeID {
			return true
		}
	}

	return false
}
