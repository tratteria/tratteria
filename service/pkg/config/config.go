package config

import (
	"fmt"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type AppConfig struct {
	TconfigdHost     string
	TconfigdSpiffeID spiffeid.ID
	MyNamespace      string
}

func GetAppConfig() (*AppConfig, error) {
	return &AppConfig{
		TconfigdHost:     getEnv("TCONFIGD_HOST"),
		TconfigdSpiffeID: spiffeid.RequireFromString(getEnv("TCONFIGD_SPIFFE_ID")),
		MyNamespace:      getEnv("MY_NAMESPACE"),
	}, nil
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}

	return value
}
