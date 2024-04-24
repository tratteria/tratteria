package authz

import (
	"net/http"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/config"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func GetSpiffeAccessControlPolicies(stocksConfig *config.StocksConfig) map[spiffeid.ID]map[string][]string {
	return map[spiffeid.ID]map[string][]string{
		stocksConfig.SpiffeIDs.Gateway: {
			"/api/stocks/search":   {http.MethodGet},
			"/api/stocks/holdings": {http.MethodGet},
			"/api/stocks/{id}":     {http.MethodGet},
		},
		stocksConfig.SpiffeIDs.Order: {
			"/internal/stocks": {http.MethodPost},
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
