package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"
)

var (
	privateKey *rsa.PrivateKey
	kid        string
	keySet     jwk.Set
)

func Initialize() error {
	return generateKeys()
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
