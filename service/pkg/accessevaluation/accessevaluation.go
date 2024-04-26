package accessevaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/oliveagle/jsonpath"
)

type AccessEvaluator struct {
	endpoint          string
	apiAuthentication apiAuthentication
	requestMapping    map[string]any
	httpClient        *http.Client
}

type apiAuthentication struct {
	method string
	token  apiAuthenticationToken
}

type apiAuthenticationToken struct {
	value string `yaml:"value"`
}

type accessEvaluationResponse struct {
	Decision bool `json:"decision"`
}

func NewAccessEvaluator(authorizationAPIconfig *config.AuthorizationAPI, httpClient *http.Client) *AccessEvaluator {
	return &AccessEvaluator{
		endpoint: authorizationAPIconfig.Endpoint,
		apiAuthentication: apiAuthentication{
			method: authorizationAPIconfig.Authentication.Method,
			token: apiAuthenticationToken{
				value: authorizationAPIconfig.Authentication.Token.Value,
			}},
		requestMapping: authorizationAPIconfig.RequestMapping,
		httpClient:     httpClient,
	}
}

func resolveJSONPaths(inputData map[string]interface{}, mapping any) (any, error) {
	switch v := mapping.(type) {
	case string:
		if strings.HasPrefix(v, "$") {
			value, err := jsonpath.JsonPathLookup(inputData, v)
			if err != nil {
				return nil, err
			}

			return value, nil
		}

		return v, nil

	case map[string]interface{}:
		resolvedMap := make(map[string]interface{})

		for key, val := range v {
			resolvedValue, err := resolveJSONPaths(inputData, val)
			if err != nil {
				continue
			}

			if resolvedValue != nil {
				resolvedMap[key] = resolvedValue
			}
		}

		return resolvedMap, nil

	default:
		return nil, fmt.Errorf("unsupported type for JSON path resolution: %T", v)
	}
}

func (a *AccessEvaluator) Evaluate(subject_token interface{}, Scope string, RequestDetails, RequestContext map[string]any) (bool, error) {
	inputData := map[string]interface{}{
		"subject_token":   subject_token,
		"scope":           Scope,
		"request_details": RequestDetails,
		"request_context": RequestContext,
	}

	requestData, err := resolveJSONPaths(inputData, a.requestMapping)
	if err != nil {
		return false, fmt.Errorf("error resolving access request mapping: %w", err)
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return false, fmt.Errorf("error marshalling access request data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error creating access request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if a.apiAuthentication.method == "Bearer" {
		req.Header.Set("Authorization", "Bearer "+a.apiAuthentication.token.value)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error sending request to the access evaluation api: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("error reading error response body from access evaluation api: %w", err)
		}
		
		return false, fmt.Errorf("access evaluation api request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response accessEvaluationResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("error decoding response from the access evaluation api: %w", err)
	}

	return response.Decision, nil
}
