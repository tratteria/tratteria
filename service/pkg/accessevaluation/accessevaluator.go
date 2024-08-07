package accessevaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tratteria/tratteria/pkg/common"
	"go.uber.org/zap"
)

type AccessEvaluationAPI struct {
	Endpoint               string         `json:"endpoint"`
	Authentication         Authentication `json:"authentication"`
	EnableAccessEvaluation bool           `json:"enableAccessEvaluation"`
}

type Authentication struct {
	Method string `json:"method"`
	Token  Token  `json:"token"`
}

type Token struct {
	Value string `json:"value"`
}

type AccessEvaluator struct {
	accessEvaluationAPI AccessEvaluationAPI
	resolvedTokenValue  string
	httpClient          *http.Client
	logger              *zap.Logger
}

type accessEvaluationResponse struct {
	Decision bool `json:"decision"`
}

func NewAccessEvaluator(accessEvaluationAPI AccessEvaluationAPI, httpClient *http.Client, logger *zap.Logger) *AccessEvaluator {
	accessEvaluator := &AccessEvaluator{
		accessEvaluationAPI: accessEvaluationAPI,
		httpClient:          httpClient,
		logger:              logger,
	}

	tokenValue := accessEvaluationAPI.Authentication.Token.Value
	if strings.HasPrefix(tokenValue, "${") && strings.HasSuffix(tokenValue, "}") {
		envVarName := strings.TrimPrefix(strings.TrimSuffix(tokenValue, "}"), "${")

		envValue := os.Getenv(envVarName)
		if envValue != "" {
			accessEvaluator.resolvedTokenValue = envValue
		} else {
			logger.Error("Environment variable %s not set", zap.String("env-var-name", envVarName))
		}
	} else {
		accessEvaluator.resolvedTokenValue = tokenValue
	}

	return accessEvaluator
}

func (ae *AccessEvaluator) Evaluate(requestMapping map[string]interface{}, subject_token interface{}, requestDetails common.RequestDetails, requestContext map[string]interface{}, pathParameter map[string]string) (bool, error) {
	if !ae.IsAccessEvaluationEnabled() {
		return true, nil
	}

	inputData := map[string]interface{}{
		"body":            requestDetails.Body,
		"headers":         requestDetails.Headers,
		"queryParameters": requestDetails.QueryParameters,
		"subject_token":   subject_token,
		"request_details": requestDetails,
		"request_context": requestContext,
	}

	for key, value := range pathParameter {
		inputData[key] = value
	}

	requestData, err := resolveJSONPaths(inputData, requestMapping)
	if err != nil {
		return false, fmt.Errorf("error resolving access request mapping: %w", err)
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return false, fmt.Errorf("error marshalling access evaluation request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, ae.accessEvaluationAPI.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error constructing access evaluation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if ae.accessEvaluationAPI.Authentication.Method == "Bearer" {
		req.Header.Set("Authorization", "Bearer "+ae.resolvedTokenValue)
	}

	resp, err := ae.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making access evaluation request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("error reading error response body from access evaluation api: %w", err)
		}

		return false, fmt.Errorf("access evaluation api request failed with non-ok status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response accessEvaluationResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("error decoding response from the access evaluation api: %w", err)
	}

	return response.Decision, nil
}

func (ae *AccessEvaluator) IsAccessEvaluationEnabled() bool {
	return ae.accessEvaluationAPI.EnableAccessEvaluation
}

func resolveJSONPaths(inputData map[string]interface{}, mapping interface{}) (interface{}, error) {
	jsonInput, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input data to JSON: %w", err)
	}

	switch v := mapping.(type) {
	case string:
		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			path := strings.TrimSuffix(strings.TrimPrefix(v, "${"), "}")

			result := gjson.GetBytes(jsonInput, path)
			if !result.Exists() {
				return nil, fmt.Errorf("failed to extract value for path %s", path)
			}

			return result.Value(), nil
		}

		return v, nil
	case map[string]interface{}:
		resolvedMap := make(map[string]interface{})

		for key, val := range v {
			resolvedValue, err := resolveJSONPaths(inputData, val)
			if err != nil {
				return nil, fmt.Errorf("error resolving value for key %s: %v", key, err)
			}

			if resolvedValue != nil {
				resolvedMap[key] = resolvedValue
			}
		}

		return resolvedMap, nil
	default:
		return v, nil
	}
}
