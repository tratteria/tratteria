package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type TokenType string

const (
	OIDC_ID_TOKEN_TYPE     TokenType = "urn:ietf:params:oauth:token-type:id_token"
	SELF_SIGNED_TOKEN_TYPE TokenType = "urn:ietf:params:oauth:token-type:self_signed"
	TXN_TOKEN_TYPE         TokenType = "urn:ietf:params:oauth:token-type:txn_token"
)

var Str2TokenType = map[string]TokenType{
	"urn:ietf:params:oauth:token-type:id_token":    OIDC_ID_TOKEN_TYPE,
	"urn:ietf:params:oauth:token-type:txn_token":   TXN_TOKEN_TYPE,
	"urn:ietf:params:oauth:token-type:self_signed": SELF_SIGNED_TOKEN_TYPE,
}

type HttpMethod string

const (
	Get     HttpMethod = http.MethodGet
	Post    HttpMethod = http.MethodPost
	Put     HttpMethod = http.MethodPut
	Delete  HttpMethod = http.MethodDelete
	Patch   HttpMethod = http.MethodPatch
	Options HttpMethod = http.MethodOptions
)

var HttpMethodList = []HttpMethod{Get, Post, Put, Delete, Patch, Options}

type RequestDetails struct {
	Path            string          `json:"endpoint"`
	Method          HttpMethod      `json:"method"`
	QueryParameters json.RawMessage `json:"queryParameters"`
	Headers         json.RawMessage `json:"headers"`
	Body            json.RawMessage `json:"body"`
}

func (r *RequestDetails) Validate() error {
	if r.Method == "" {
		return errors.New("method cannot be empty")
	}

	if r.Path == "" {
		return errors.New("endpoint cannot be empty")
	}

	if parsedPath, err := url.Parse(r.Path); err != nil {
		return fmt.Errorf("invalid endpoint: %v", err)
	} else if parsedPath.Path != r.Path {
		return errors.New("endpoint must not include domain or scheme")
	}

	return nil
}

type TokenRequest struct {
	Audience           string
	RequestedTokenType TokenType
	SubjectToken       string
	SubjectTokenType   TokenType
	RequestDetails     RequestDetails
	RequestContext     map[string]any
}
