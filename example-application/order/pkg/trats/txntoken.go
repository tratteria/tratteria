package trats

import "github.com/golang-jwt/jwt"

type TxnToken struct {
	TransactionID        string           `json:"txn"`
	Subject              Subject          `json:"sub"`
	RequesterContext     requesterContext `json:"req_ctx"`
	Purpose              string           `json:"purp"`
	AuthorizationContext any              `json:"azd"`
	jwt.StandardClaims
}

type requesterContext struct {
	RequesterIP          string `json:"req_ip"`
	AuthenticationMethod string `json:"authn"`
	RequestingWorkload   string `json:"req_wl"`
}

type Subject struct {
	Format string `json:"format"`
	Email  string `json:"email"`
}
