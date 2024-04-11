package config

import (
	"context"
	"fmt"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
)

type spiffeIDs struct {
	Gateway spiffeid.ID
	Order   spiffeid.ID
	Stocks  spiffeid.ID
}

type StocksConfig struct {
	SpiffeIDs *spiffeIDs
}

func GetStocksConfig() *StocksConfig {
	return &StocksConfig{
		SpiffeIDs: &spiffeIDs{
			Gateway: spiffeid.RequireFromString(getEnv("GATEWAY_SERVICE_SPIFFE_ID")),
			Order:   spiffeid.RequireFromString(getEnv("ORDER_SERVICE_SPIFFE_ID")),
			Stocks:  spiffeid.RequireFromString(getEnv("STOCKS_SERVICE_SPIFFE_ID")),
		},
	}
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
