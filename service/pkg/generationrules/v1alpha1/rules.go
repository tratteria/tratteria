package v1alpha1

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/common"

	"errors"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

type GenerationRulesManager interface {
	AddRule(rule GenerationRule)
	GetRules() map[string]map[string]GenerationRule
	GetRulesVersionId() string
}

type GenerationRulesApplier interface {
	ApplyRule(path string, method common.HttpMethod, input map[string]interface{}) (string, map[string]interface{}, error)
}

type GenerationRule struct {
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

type AzdMapping map[string]AzdField
type AzdField struct {
	Value string `json:"value"`
}

type GenerationRules struct {
	rules          map[string]map[string]GenerationRule
	rulesVersionId string
	mu             sync.RWMutex
}

func NewGenerationRules() *GenerationRules {
	return &GenerationRules{
		rules: make(map[string]map[string]GenerationRule),
	}
}

func (m *GenerationRules) AddRule(rule GenerationRule) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exist := m.rules[rule.Method]; !exist {
		m.rules[rule.Method] = make(map[string]GenerationRule)
	}

	m.rules[rule.Method][rule.Endpoint] = rule
}

func (m *GenerationRules) GetRules() map[string]map[string]GenerationRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.rules
}

// Read lock should be take by the function calling matchRule.
func (m *GenerationRules) matchRule(path string, method common.HttpMethod) (GenerationRule, map[string]string, error) {
	methodRuleMap, ok := m.rules[string(method)]
	if !ok {
		return GenerationRule{}, nil, fmt.Errorf("no rules found for %s HTTP method", string(method))
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

	return GenerationRule{}, nil, errors.New("no matching rule found")
}

func convertToRegex(template string) string {
	r := strings.NewReplacer("{#", "(?P<", "}", ">[^/]+)")

	return "^" + r.Replace(template) + "$"
}

func (m *GenerationRules) GetRulesVersionId() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.rulesVersionId
}

func (m *GenerationRules) ApplyRule(path string, method common.HttpMethod, input map[string]interface{}) (string, map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	generationRule, pathParameter, err := m.matchRule(path, method)
	if err != nil {
		return "", nil, fmt.Errorf("error matching generation rule for %s path and %s method: %w", path, string(method), err)
	}

	for par, val := range pathParameter {
		input[par] = val
	}

	azd, err := m.computeAzd(generationRule.AzdMapping, input)
	if err != nil {
		return "", nil, fmt.Errorf("error computing azd from generation rule for %s path and %s method: %w", path, string(method), err)
	}

	return generationRule.Purp, azd, nil
}

// Read lock should be take by the function calling computeAzd.
func (m *GenerationRules) computeAzd(azdMapping AzdMapping, input map[string]interface{}) (map[string]interface{}, error) {
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
