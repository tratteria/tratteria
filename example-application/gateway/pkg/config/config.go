package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const OIDC_PROVIDER_INITILIZATION_MAX_RETRIES = 5

type spiffeIDs struct {
	TxnToken spiffeid.ID
	Gateway  spiffeid.ID
	Order    spiffeid.ID
	Stocks   spiffeid.ID
}

type GatewayConfig struct {
	TxnTokenServiceURL string
	StocksServiceURL   string
	OrderServiceURL    string
	SpiffeIDs          *spiffeIDs
}

func GetGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		TxnTokenServiceURL: getEnv("TXN_TOKEN_SERVICE_URL"),
		StocksServiceURL:   getEnv("STOCKS_SERVICE_URL"),
		OrderServiceURL:    getEnv("ORDER_SERVICE_URL"),
		SpiffeIDs: &spiffeIDs{
			TxnToken: spiffeid.RequireFromString(getEnv("TXN_TOKEN_SERVICE_SPIFFE_ID")),
			Gateway:  spiffeid.RequireFromString(getEnv("GATEWAY_SERVICE_SPIFFE_ID")),
			Order:    spiffeid.RequireFromString(getEnv("ORDER_SERVICE_SPIFFE_ID")),
			Stocks:   spiffeid.RequireFromString(getEnv("STOCKS_SERVICE_SPIFFE_ID")),
		},
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

		logger.Error("Failed to connect to the OIDC provider.",
		zap.Int("attempt", i+1),
		zap.String("retrying_in", delay.String()),
		zap.Error(err))
		time.Sleep(delay)

		delay *= 2
	}

	logger.Error(fmt.Sprintf("Failed to connect to the OIDC provider after %d attempts", OIDC_PROVIDER_INITILIZATION_MAX_RETRIES))

	panic(fmt.Sprintf("failed to connect to the OIDC provider after %d attempts", OIDC_PROVIDER_INITILIZATION_MAX_RETRIES))
}

func GetSpireJwtSource(logger *zap.Logger) *workloadapi.JWTSource {
	ctx := context.Background()

	jwtSource, err := workloadapi.NewJWTSource(ctx)
	if err != nil {
		logger.Fatal("Unable to create SPIRE JWTSource for fetching JWT-SVIDs.", zap.Error(err))
	}

	logger.Info("Successfully created SPIRE JWTSource for fetching JWT-SVIDs.")

	return jwtSource
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}

	return value
}
