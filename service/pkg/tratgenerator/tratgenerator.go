package tratgenerator

import (
	"encoding/json"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/common"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/generationrules/v1alpha1"
)

type TraTGenerator struct {
	generationRulesMatcher v1alpha1.GenerationRulesApplier
}

func NewTraTGenerator(generationRulesMatcher v1alpha1.GenerationRulesApplier) *TraTGenerator {
	return &TraTGenerator{
		generationRulesMatcher: generationRulesMatcher,
	}
}

func (tv *TraTGenerator) GenerateTraT(path string, method common.HttpMethod, queryParameters json.RawMessage, headers json.RawMessage, body json.RawMessage) (string, map[string]any, error) {
	input := make(map[string]interface{})

	input["body"] = body
	input["headers"] = headers
	input["queryParameters"] = queryParameters

	return tv.generationRulesMatcher.ApplyRule(path, method, input)
}
