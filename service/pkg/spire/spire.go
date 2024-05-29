package spire

import (
	"context"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/tratteria/tratteria/pkg/config"
	"go.uber.org/zap"
)

func NewSpireJwtSource(endpointSocket string) (*workloadapi.JWTSource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jwtSource, err := workloadapi.NewJWTSource(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(endpointSocket)))
	if err != nil {
		return nil, err
	}

	return jwtSource, nil
}

func GetSpireJwtSource(appConfig *config.AppConfig, logger *zap.Logger) (*workloadapi.JWTSource, error) {
	if appConfig.Spiffe == nil {
		logger.Warn("SPIFFE is not configured; ensure your architecture securely authenticates requesting workloads using similar methods. Avoid insecure mechanisms such as long-lived shared secrets.")

		return nil, nil
	}

	spireJwtSource, err := NewSpireJwtSource(appConfig.Spiffe.EndpointSocket)
	if err != nil {
		return nil, err
	}

	return spireJwtSource, nil
}
