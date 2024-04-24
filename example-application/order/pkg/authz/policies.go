package authz

import (
	"net/http"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/config"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func GetSpiffeAccessControlPolicies(orderConfig *config.OrderConfig) map[spiffeid.ID]map[string][]string {
	return map[spiffeid.ID]map[string][]string{
		orderConfig.SpiffeIDs.Gateway: {
			"/api/order":      {http.MethodPost},
			"/api/order/{id}": {http.MethodGet},
		},
	}
}

func IsSpiffeIDAuthorized(spiffeID spiffeid.ID, method, path string, policies map[spiffeid.ID]map[string][]string) bool {
	allowedPaths, ok := policies[spiffeID]
	if !ok {
		return false
	}

	allowedMethods, ok := allowedPaths[path]
	if !ok {
		return false
	}

	for _, m := range allowedMethods {
		if m == method {
			return true
		}
	}

	return false
}
