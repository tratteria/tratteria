package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/trats"
)

type spiffeIDs struct {
	Gateway spiffeid.ID
	Order   spiffeid.ID
	Stocks  spiffeid.ID
}

type txnTokenKeys struct {
	JWKS string
}

type toggles struct {
	SpireToggle    bool
	TxnTokenToggle bool
}

type StocksConfig struct {
	SpiffeIDs    *spiffeIDs
	TxnTokenKeys *txnTokenKeys
	Toggles      *toggles
}

func GetStocksConfig() *StocksConfig {
	return &StocksConfig{
		SpiffeIDs: &spiffeIDs{
			Gateway: spiffeid.RequireFromString(getEnv("GATEWAY_SERVICE_SPIFFE_ID")),
			Order:   spiffeid.RequireFromString(getEnv("ORDER_SERVICE_SPIFFE_ID")),
			Stocks:  spiffeid.RequireFromString(getEnv("STOCKS_SERVICE_SPIFFE_ID")),
		},
		TxnTokenKeys: &txnTokenKeys{
			JWKS: getEnv("TTS_JWKS"),
		},
		Toggles: &toggles{
			SpireToggle:    getBoolEnv("ENABLE_SPIRE"),
			TxnTokenToggle: getBoolEnv("ENABLE_TXN_TOKEN"),
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

func GetTraTsVerifier() *trats.Verifier {
	return trats.NewVerifier(getEnv("TRATS_AUDIENCE"),
		getEnv("TRATS_ISSUER"))
}

func getBoolEnv(key string) bool {
	val, err := strconv.ParseBool(getEnv(key))
	if err != nil {
		panic("Error parsing boolean environment variable " + key + ": " + err.Error())
	}

	return val
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}

	return value
}
