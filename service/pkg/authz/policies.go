package authz

import (
	"fmt"
	"net/http"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/generationrules/v1alpha1"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func GetPublicEndpoints() []string {
	return []string{
		"/.well-known/jwks.json",
		"/generation-rules",
		"/generation-endpoint-rule-webhook", // TODO: should be protected with mTLS
		"/generation-token-rule-webhook",    // TODO: should be protected with mTLS
	}
}

func GetSpiffeAccessControlPolicies(generationRules *v1alpha1.GenerationRulesImp) (map[string]map[string][]spiffeid.ID, error) {
	tokenAuthorizedIds, err := generationRules.GetAuthorizedSpifeeIDs()
	if err != nil {
		return map[string]map[string][]spiffeid.ID{}, fmt.Errorf("error getting token authorized spiffe ids: %w", err)
	}

	return map[string]map[string][]spiffeid.ID{
		"/token_endpoint": {http.MethodPost: tokenAuthorizedIds},
	}, nil
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
