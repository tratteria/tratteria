package accessevaluation

import (
	"encoding/json"
	"net/http"
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
	accessEvaluationAPI *AccessEvaluationAPI
	httpClient          *http.Client
}

func NewAccessEvaluator(accessEvaluationAPI *AccessEvaluationAPI, httpClient *http.Client) *AccessEvaluator {
	return &AccessEvaluator{
		accessEvaluationAPI: accessEvaluationAPI,
		httpClient:          httpClient,
	}
}

func (a *AccessEvaluator) Evaluate(accessEvalationRequest json.RawMessage) (bool, error) {
	//TODO
	return true, nil
}

func (a *AccessEvaluator) IsAccessEvaluationEnabled() bool {
	if a.accessEvaluationAPI == nil {
		return false
	}

	return a.accessEvaluationAPI.EnableAccessEvaluation
}
