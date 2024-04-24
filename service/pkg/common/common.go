package common

type TokenType string

const (
	OIDC_ID_TOKEN_TYPE TokenType = "urn:ietf:params:oauth:token-type:id_token"
	TXN_TOKEN_TYPE     TokenType = "urn:ietf:params:oauth:token-type:txn_token"
)

var Str2TokenType = map[string]TokenType{
	"urn:ietf:params:oauth:token-type:id_token":  OIDC_ID_TOKEN_TYPE,
	"urn:ietf:params:oauth:token-type:txn_token": TXN_TOKEN_TYPE,
}
