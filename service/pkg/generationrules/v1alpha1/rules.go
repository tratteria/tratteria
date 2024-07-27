package v1alpha1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/accessevaluation"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/common"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/subjecttokenhandler"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"go.uber.org/zap"

	"errors"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

type TratteriaConfigToken struct {
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
	LifeTime string `json:"lifeTime"`
}

type TratteriaConfigGenerationRule struct {
	Token                               TratteriaConfigToken                  `json:"token"`
	SubjectTokens                       *subjecttokenhandler.SubjectTokens    `json:"subjectTokens"`
	AccessEvaluationAPI                 *accessevaluation.AccessEvaluationAPI `json:"accessEvaluationAPI"`
	TokenGenerationAuthorizedServiceIds []string                              `json:"tokenGenerationAuthorizedServiceIds"`
}

type TraTGenerationRule struct {
	Endpoint   string            `json:"endpoint"`
	Method     common.HttpMethod `json:"method"`
	Purp       string            `json:"purp"`
	AzdMapping AzdMapping        `json:"azdmapping,omitempty"`
}

type AzdMapping map[string]AzdField
type AzdField struct {
	Value string `json:"value"`
}

type TraTGenerationRules map[common.HttpMethod]map[string]TraTGenerationRule

type GenerationRules struct {
	TratteriaConfigRules *TratteriaConfigGenerationRule `json:"tratteriaConfigRules"`
	TraTRules            TraTGenerationRules            `json:"traTRules"`
}

func NewGenerationRules() *GenerationRules {
	traTRules := make(TraTGenerationRules)

	for _, method := range common.HttpMethodList {
		traTRules[method] = make(map[string]TraTGenerationRule)
	}

	return &GenerationRules{
		TraTRules: traTRules,
	}
}

type GenerationRulesImp struct {
	rules                *GenerationRules
	subjectTokenHandlers *subjecttokenhandler.TokenHandlers
	accessevaluator      *accessevaluation.AccessEvaluator
	httpClient           *http.Client
	logger               *zap.Logger
	mu                   sync.RWMutex
}

func NewGenerationRulesImp(httpClient *http.Client, logger *zap.Logger) *GenerationRulesImp {
	return &GenerationRulesImp{
		rules:      NewGenerationRules(),
		httpClient: httpClient,
		logger:     logger,
	}
}

func (gri *GenerationRulesImp) AddTraTRule(generationTraTRule TraTGenerationRule) error {
	gri.mu.Lock()
	defer gri.mu.Unlock()

	if _, exist := gri.rules.TraTRules[generationTraTRule.Method]; !exist {
		return fmt.Errorf("invalid HTTP method: %s", string(generationTraTRule.Method))
	}

	gri.rules.TraTRules[generationTraTRule.Method][generationTraTRule.Endpoint] = generationTraTRule

	return nil
}

func (gri *GenerationRulesImp) UpdateTratteriaConfigRule(generationTratteriaConfigRule TratteriaConfigGenerationRule) {
	gri.mu.Lock()
	defer gri.mu.Unlock()

	gri.rules.TratteriaConfigRules = &generationTratteriaConfigRule
	gri.subjectTokenHandlers = subjecttokenhandler.NewTokenHandlers(generationTratteriaConfigRule.SubjectTokens, gri.logger)
	gri.accessevaluator = accessevaluation.NewAccessEvaluator(generationTratteriaConfigRule.AccessEvaluationAPI, gri.httpClient)
}

func (gri *GenerationRulesImp) GetRulesJSON() (json.RawMessage, error) {
	gri.mu.RLock()
	defer gri.mu.RUnlock()

	jsonData, err := json.Marshal(gri.rules)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// Read lock should be take by the function calling matchRule.
func (gri *GenerationRulesImp) matchRule(path string, method common.HttpMethod) (TraTGenerationRule, map[string]string, error) {
	methodRuleMap, ok := gri.rules.TraTRules[method]
	if !ok {
		return TraTGenerationRule{}, nil, fmt.Errorf("invalid HTTP method: %s", string(method))
	}

	for pattern, rule := range methodRuleMap {
		regexPattern := convertToRegex(pattern)
		re := regexp.MustCompile(regexPattern)

		if re.MatchString(path) {
			matches := re.FindStringSubmatch(path)
			names := re.SubexpNames()

			pathParameters := make(map[string]string)

			for i, name := range names {
				if i != 0 && name != "" {
					pathParameters[name] = matches[i]
				}
			}

			return rule, pathParameters, nil
		}
	}

	return TraTGenerationRule{}, nil, errors.New("no matching rule found")
}

func convertToRegex(template string) string {
	r := strings.NewReplacer("{#", "(?P<", "}", ">[^/]+)")

	return "^" + r.Replace(template) + "$"
}

func (gri *GenerationRulesImp) ConstructScopeAndAzd(txnTokenRequest *common.TokenRequest) (string, map[string]interface{}, error) {
	gri.mu.RLock()
	defer gri.mu.RUnlock()

	path := txnTokenRequest.RequestDetails.Path
	method := txnTokenRequest.RequestDetails.Method

	input := make(map[string]interface{})
	input["body"] = txnTokenRequest.RequestDetails.Body
	input["headers"] = txnTokenRequest.RequestDetails.Headers
	input["queryParameters"] = txnTokenRequest.RequestDetails.QueryParameters

	generationTraTRule, pathParameter, err := gri.matchRule(path, method)
	if err != nil {
		return "", nil, fmt.Errorf("error matching generation rule for %s path and %s method: %w", path, string(method), err)
	}

	for par, val := range pathParameter {
		input[par] = val
	}

	if generationTraTRule.AzdMapping == nil {
		return generationTraTRule.Purp, nil, nil
	}

	azd, err := gri.computeAzd(generationTraTRule.AzdMapping, input)
	if err != nil {
		return "", nil, fmt.Errorf("error computing azd from generation trat rule for %s path and %s method: %w", path, string(method), err)
	}

	return generationTraTRule.Purp, azd, nil
}

// Read lock should be take by the function calling computeAzd.
func (gri *GenerationRulesImp) computeAzd(azdMapping AzdMapping, input map[string]interface{}) (map[string]interface{}, error) {
	azd := make(map[string]interface{})

	jsonInput, err := marshalToJson(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input to JSON: %w", err)
	}

	for key, azdField := range azdMapping {
		valueSpec := azdField.Value

		if strings.HasPrefix(valueSpec, "${") && strings.HasSuffix(valueSpec, "}") {
			path := strings.TrimSuffix(strings.TrimPrefix(valueSpec, "${"), "}")
			value := extractValueFromJson(jsonInput, path)

			if value == nil {
				return nil, fmt.Errorf("failed to extract value for key %s from path %s", key, path)
			}

			azd[key] = value
		} else {
			azd[key] = valueSpec
		}
	}

	return azd, nil
}

func extractValueFromJson(jsonStr string, path string) interface{} {
	result := gjson.Get(jsonStr, path)

	if result.Exists() {
		return result.Value()
	}

	return nil
}

func marshalToJson(data map[string]interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (gri *GenerationRulesImp) GetIssuer() string {
	gri.mu.RLock()
	defer gri.mu.RUnlock()

	return gri.rules.TratteriaConfigRules.Token.Issuer
}

func (gri *GenerationRulesImp) GetAudience() string {
	gri.mu.RLock()
	defer gri.mu.RUnlock()

	return gri.rules.TratteriaConfigRules.Token.Audience
}

func (gri *GenerationRulesImp) GetTokenLifetime() (time.Duration, error) {
	gri.mu.RLock()
	defer gri.mu.RUnlock()

	duration, err := time.ParseDuration(gri.rules.TratteriaConfigRules.Token.LifeTime)
	if err != nil {
		return 0, fmt.Errorf("error parsing token lifetime: %v", err)
	}

	return duration, nil
}

func (gri *GenerationRulesImp) EvaluateAccess(txnTokenRequest *common.TokenRequest, subjectTokenClaims interface{}, scope string, azd map[string]any) (bool, error) {
	gri.mu.RLock()
	defer gri.mu.RUnlock()

	if !gri.accessevaluator.IsAccessEvaluationEnabled() {
		return true, nil
	}

	// TODO: implemente access evaluation

	return true, nil
}

func (gri *GenerationRulesImp) GetSubjectTokenHandler(tokenType common.TokenType) (subjecttokenhandler.TokenHandler, error) {
	return gri.subjectTokenHandlers.GetHandler(tokenType)
}

func (gri *GenerationRulesImp) GetTraTGenerationAuthorizedServicesSpifeeIDs() ([]spiffeid.ID, error) {
	if gri.rules.TratteriaConfigRules == nil {
		return []spiffeid.ID{}, nil
	}

	stringIDs := gri.rules.TratteriaConfigRules.TokenGenerationAuthorizedServiceIds
	spiffeIDs := make([]spiffeid.ID, 0, len(stringIDs))

	for _, idStr := range stringIDs {
		id, err := spiffeid.FromString(idStr)
		if err != nil {
			return nil, err
		}

		spiffeIDs = append(spiffeIDs, id)
	}

	return spiffeIDs, nil
}

type TconfigdGenerationRules struct {
	TratteriaConfigGenerationRule *TratteriaConfigGenerationRule `json:"tratteriaConfigGenerationRule"`
	TratGenerationRules           []*TraTGenerationRule          `json:"tratGenerationRules"`
}

func (gri *GenerationRulesImp) UpdateCompleteRules(tconfigdGenerationRules *TconfigdGenerationRules) {
	gri.mu.Lock()
	defer gri.mu.Unlock()

	gri.rules.TratteriaConfigRules = tconfigdGenerationRules.TratteriaConfigGenerationRule

	if gri.rules.TratteriaConfigRules != nil {
		gri.subjectTokenHandlers = subjecttokenhandler.NewTokenHandlers(gri.rules.TratteriaConfigRules.SubjectTokens, gri.logger)
		gri.accessevaluator = accessevaluation.NewAccessEvaluator(gri.rules.TratteriaConfigRules.AccessEvaluationAPI, gri.httpClient)
	}

	traTRules := make(TraTGenerationRules)

	for _, method := range common.HttpMethodList {
		traTRules[method] = make(map[string]TraTGenerationRule)
	}

	for _, endpointRule := range tconfigdGenerationRules.TratGenerationRules {
		traTRules[endpointRule.Method][endpointRule.Endpoint] = *endpointRule
	}

	gri.rules.TraTRules = traTRules
}
