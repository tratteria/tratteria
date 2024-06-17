package common

import (
	"errors"
	"fmt"
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
	Get     HttpMethod = "GET"
	Post    HttpMethod = "POST"
	Put     HttpMethod = "PUT"
	Delete  HttpMethod = "DELETE"
	Patch   HttpMethod = "PATCH"
	Options HttpMethod = "OPTIONS"
)

type RequestDetails struct {
	Endpoint        string                 `json:"endpoint"`
	Method          HttpMethod             `json:"method"`
	Body            map[string]interface{} `json:"body"`
	Headers         map[string]interface{} `json:"headers"`
	QueryParameters map[string]interface{} `json:"queryParameters"`
}

func (r *RequestDetails) Validate() error {
	if r.Method == "" {
		return errors.New("method cannot be empty")
	}

	if r.Endpoint == "" {
		return errors.New("path cannot be empty")
	}

	if parsedPath, err := url.Parse(r.Endpoint); err != nil {
		return fmt.Errorf("invalid path: %v", err)
	} else if parsedPath.Path != r.Endpoint {
		return errors.New("path must not include domain or scheme")
	}

	return nil
}

type TokenRequest struct {
	RequestedTokenType TokenType
	SubjectToken       string
	SubjectTokenType   TokenType
	RequestDetails     RequestDetails
	RequestContext     map[string]any
}
