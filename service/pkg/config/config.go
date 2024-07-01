package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type AppConfig struct {
	TconfigdUrl url.URL
	SpiffeID    spiffeid.ID
	MyNamespace string
}

func GetAppConfig() (*AppConfig, error) {
	tconfigdUrl, err := url.Parse(getEnv("TCONFIGD_URL"))
	if err != nil {
		return nil, fmt.Errorf("error parsing tconfigd url from environment variable: %w", err)
	}
	return &AppConfig{
		TconfigdUrl: *tconfigdUrl,
		SpiffeID:    spiffeid.RequireFromString(getEnv("SPIFFE_ID")),
		MyNamespace: getEnv("MY_NAMESPACE"),
	}, nil
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}

	return value
}
