package trats

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
)

type Verifier struct {
	Audience string
	Issuer   string
}

func NewVerifier(audience string, issuer string) *Verifier {
	return &Verifier{
		Audience: audience,
		Issuer:   issuer,
	}
}

func (v *Verifier) ParseAndVerify(tokenStr string, jwksJson string) (*TxnToken, error) {
	keySet, err := jwk.Parse([]byte(jwksJson))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenStr, &TxnToken{}, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		keys, found := keySet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("key %v not found in JWKS", kid)
		}

		var rawKey interface{}
		if err := keys.Raw(&rawKey); err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		switch key := rawKey.(type) {
		case *rsa.PublicKey:
			return key, nil
		case *ecdsa.PublicKey:
			return key, nil
		default:
			return nil, fmt.Errorf("unsupported key type %T", rawKey)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*TxnToken)
	if !ok {
		return nil, fmt.Errorf("failed to extract token claims")
	}

	if !claims.VerifyAudience(v.Audience, true) {
		return nil, fmt.Errorf("invalid audience: %v", claims.Audience)
	}

	if !claims.VerifyIssuer(v.Issuer, true) {
		return nil, fmt.Errorf("invalid issuer: %v", claims.Issuer)
	}

	return claims, nil
}

// ParseWithoutVerification parses the transaction token JWT but does not verify it.
// WARNING: Using this function means that the integrity and authenticity of the token are not verified.
// This should only be used in contexts where verification is explicitly not required or has been handled elsewhere.
func (v *Verifier) ParseWithoutVerification(tokenStr string) (*TxnToken, error) {
	log.Println("WARNING: Parsing token without verification. This should not be used in production environments where security is a concern.")

	// The nil function passed as the third parameter indicates no verification of the signature.
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, &TxnToken{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TxnToken)
	if !ok {
		return nil, fmt.Errorf("failed to extract token claims")
	}

	return claims, nil
}
