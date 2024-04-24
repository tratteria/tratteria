package main

import (
	"log"
	"net/http"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/handler"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/config"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Cannot initialize Zap logger: %v.", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error syncing logger: %v", err)
		}
	}()

	gatewayConfig := config.GetGatewayConfig()
	
	oauth2Config := config.GetOauth2Config()
	oidcProvider := config.GetOIDCProvider(logger)
	
	spireJwtSource := config.GetSpireJwtSource(logger)

	defer spireJwtSource.Close()

	httpClient := &http.Client{}

	router := handler.SetupRoutes(gatewayConfig, oauth2Config, oidcProvider, spireJwtSource, httpClient, logger)

	srv := &http.Server{
		Addr:         "0.0.0.0:30000",
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on 30000.")
	log.Fatal(srv.ListenAndServe())
}
