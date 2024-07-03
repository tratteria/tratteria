package spire

import (
	"context"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

const JWTSourceTimeout = 15 * time.Second

func NewSpireJwtSource(endpointSocket string) (*workloadapi.JWTSource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jwtSource, err := workloadapi.NewJWTSource(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(endpointSocket)))
	if err != nil {
		return nil, err
	}

	return jwtSource, nil
}

func GetSpireJwtSource() (*workloadapi.JWTSource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), JWTSourceTimeout)
	defer cancel()

	jwtSource, err := workloadapi.NewJWTSource(ctx)
	if err != nil {
		return nil, err
	}

	return jwtSource, nil
}
