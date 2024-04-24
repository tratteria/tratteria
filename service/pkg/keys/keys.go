package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
)

type JWKS struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

var (
	privateKey *rsa.PrivateKey
	jwks       JWKS
	kid        string
)

func Initialize(appConfig *config.AppConfig) error {
	kid = appConfig.Keys.KeyID

	err := json.Unmarshal([]byte(appConfig.Keys.JWKS), &jwks)
	if err != nil {
		return fmt.Errorf("error unmarshalling JWKS: %w", err)
	}

	privateKeyPem, err := base64.StdEncoding.DecodeString(appConfig.Keys.PrivateKey)
	if err != nil {
		return fmt.Errorf("error decoding base64 private key: %w", err)
	}

	block, rest := pem.Decode(privateKeyPem)
	if block == nil {
		return errors.New("failed to decode private key PEM block")
	}

	if len(rest) > 0 {
		return errors.New("unexpected extra data found after PEM block")
	}

	privInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("error parsing private key: %w", err)
	}

	var ok bool
	if privateKey, ok = privInterface.(*rsa.PrivateKey); !ok {
		return fmt.Errorf("failed to assert type to *rsa.PrivateKey")
	}

	return nil
}

func GetPrivateKey() *rsa.PrivateKey {
	return privateKey
}

func GetKid() string {
	return kid
}

func GetJWKS() JWKS {
	return jwks
}
