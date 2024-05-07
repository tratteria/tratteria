package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/lestrrat-go/jwx/jwk"
)

var (
	privateKey *rsa.PrivateKey
	kid        string
	keySet     jwk.Set
)

func Initialize(appConfig *config.AppConfig) error {
	if appConfig.Keys == nil {
		return generateKeys()
	} else {
		return parseKeys(appConfig)
	}
}

func parseKeys(appConfig *config.AppConfig) error {
	var err error
	kid = appConfig.Keys.KeyID

	keySet, err = jwk.Parse([]byte(appConfig.Keys.JWKS))
	if err != nil {
		return fmt.Errorf("error parsing JWKS: %w", err)
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

func generateKeys() error {
	var err error

	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	publicKey := privateKey.PublicKey
	
	jwkKey, err := jwk.New(&publicKey)
	if err != nil {
		return fmt.Errorf("failed to create JWK from public key: %w", err)
	}

	if err := jwkKey.Set(jwk.AlgorithmKey, "RS256"); err != nil {
		return fmt.Errorf("failed to set algorithm for JWK: %w", err)
	}

	if err := jwkKey.Set(jwk.KeyUsageKey, "sig"); err != nil {
		return fmt.Errorf("failed to set usage for JWK: %w", err)
	}

	kidBytes := make([]byte, 16)
	if _, err := rand.Read(kidBytes); err != nil {
		return fmt.Errorf("failed to generate random bytes for Key ID: %w", err)
	}
	
	kid = base64.StdEncoding.EncodeToString(kidBytes)

	if err := jwkKey.Set(jwk.KeyIDKey, kid); err != nil {
		return fmt.Errorf("failed to set Key ID for JWK: %w", err)
	}

	keySet = jwk.NewSet()
	
	keySet.Add(jwkKey)

	return nil
}

func GetPrivateKey() *rsa.PrivateKey {
	return privateKey
}

func GetKid() string {
	return kid
}

func GetJWKS() jwk.Set {
	return keySet
}
