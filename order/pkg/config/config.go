package config

import (
	"fmt"
	"os"
)

type OrderConfig struct {
	StocksServiceURL string
}

func NewConfig() *OrderConfig {
	return &OrderConfig{
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
