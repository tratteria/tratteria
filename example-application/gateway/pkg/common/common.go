package common

type IDTokenClaims struct {
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}

type contextKey string

const OIDC_ID_TOKEN_CONTEXT_KEY contextKey = "id_token"
