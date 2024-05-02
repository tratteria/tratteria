package common

import (
	"net/url"
	"path"
)

type IDTokenClaims struct {
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}

type contextKey string

const OIDC_ID_TOKEN_CONTEXT_KEY contextKey = "id_token"

func AppendPathToURL(baseURL *url.URL, appendPath string) *url.URL {
	newURL := *baseURL
	newURL.Path = path.Join(newURL.Path, appendPath)
	return &newURL
}
