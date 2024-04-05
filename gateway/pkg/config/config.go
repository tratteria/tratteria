package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const OIDC_PROVIDER_INITILIZATION_MAX_RETRIES = 5

type GatewayConfig struct {
	StocksServiceURL string
	OrderServiceURL  string
}

func GetGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		StocksServiceURL: getEnv("STOCKS_SERVICE_URL"),
		OrderServiceURL:  getEnv("ORDER_SERVICE_URL"),
	}
}

func GetOauth2Config() oauth2.Config {
	return oauth2.Config{
		ClientID:     getEnv("OAUTH2_CLIENT_ID"),
		ClientSecret: getEnv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  getEnv("OAUTH2_REDIRECT_URL"),
		Endpoint: oauth2.Endpoint{
			TokenURL: getEnv("OAUTH2_TOKEN_URL"),
		},
		Scopes: []string{"openid", "profile", "email"},
	}
}

func GetOIDCProvider(logger *zap.Logger) *oidc.Provider {
	delay := time.Second

	for i := 0; i < OIDC_PROVIDER_INITILIZATION_MAX_RETRIES; i++ {
		ctx := context.Background()
		oidcIssuer := getEnv("OIDC_ISSUER_URL")

		provider, err := oidc.NewProvider(ctx, oidcIssuer)
		if err == nil {
			logger.Info("Successfully connected to the OIDC provider.")

			return provider
		}

		logger.Error(fmt.Sprintf("Failed to connect to the OIDC provider, retrying in %v...\n", delay))
		time.Sleep(delay)
		
		delay *= 2
	}

	logger.Error(fmt.Sprintf("Failed to connect to the OIDC provider after %d attempts", OIDC_PROVIDER_INITILIZATION_MAX_RETRIES))
	
	panic(fmt.Sprintf("failed to connect to the OIDC provider after %d attempts", OIDC_PROVIDER_INITILIZATION_MAX_RETRIES))
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}
	
	return value
}
