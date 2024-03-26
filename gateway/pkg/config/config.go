package config

import (
	"fmt"
	"os"
)

type GatewayConfig struct {
	StocksServiceURL string
}

func NewConfig() *GatewayConfig {
	return &GatewayConfig{
		StocksServiceURL: getEnv("STOCKS_SERVICE_URL"),
	}
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}
	return value
}
