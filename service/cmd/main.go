package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/handler"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/configsync"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/generationrules/v1alpha1"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/keys"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/middlewares"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/service"
)

const (
	HTTPS_PORT           = 443
	HTTP_PORT            = 80
	SPIRE_SOURCE_TIMEOUT = 15 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Cannot initialize Zap logger: %v.", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error syncing logger: %v", err)
		}
	}()

	x509SrcCtx, cancel := context.WithTimeout(context.Background(), SPIRE_SOURCE_TIMEOUT)

	defer cancel()

	x509Source, err := workloadapi.NewX509Source(x509SrcCtx)
	if err != nil {
		logger.Fatal("Failed to create SPIRE X.509 source", zap.Error(err))
	}

	defer x509Source.Close()

	appConfig, err := config.GetAppConfig()
	if err != nil {
		logger.Fatal("Error getting application configuration.", zap.Error(err))
	}

	err = keys.Initialize()
	if err != nil {
		logger.Fatal("Error initializing keys:", zap.Error(err))
	}

	httpClient := &http.Client{}
	generationRules := v1alpha1.NewGenerationRulesImp(httpClient, logger)

	configSyncClient, err := configsync.NewClient(HTTPS_PORT, appConfig.TconfigdUrl, appConfig.TconfigdSpiffeID, appConfig.MyNamespace, generationRules, x509Source, logger)
	if err != nil {
		logger.Fatal("Error creating configuration sync client for tconfigd", zap.Error(err))
	}

	if err := configSyncClient.Start(); err != nil {
		logger.Fatal("Error establishing communication with tconfigd", zap.Error(err))
	}

	appService := service.NewService(generationRules, logger)
	appHandler := handler.NewHandlers(appService, logger)

	go func() {
		err := startHTTPServer(appHandler, logger)
		if err != nil {
			logger.Fatal("HTTP server exited with error", zap.Error(err))
		}
	}()

	tconfigdSpiffeID := func() ([]spiffeid.ID, error) { return []spiffeid.ID{appConfig.TconfigdSpiffeID}, nil }
	traTGenAuthorizedSpiffeIDs := func() ([]spiffeid.ID, error) { return generationRules.GetTraTGenerationAuthorizedServicesSpifeeIDs() }

	go func() {
		if err := startHTTPSServer(
			appHandler,
			x509Source,
			tconfigdSpiffeID,
			traTGenAuthorizedSpiffeIDs,
			logger,
		); err != nil {
			logger.Fatal("HTTPS server exited with error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Info("Shutting down tratteria...")
}

func startHTTPServer(handlers *handler.Handlers, logger *zap.Logger) error {
	router := mux.NewRouter()
	router.HandleFunc("/generation-rules", handlers.GetGenerationRulesHandler).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", HTTP_PORT),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	logger.Info("Starting HTTP server...", zap.Int("port", HTTP_PORT))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Failed to start the http api server", zap.Error(err))

		return fmt.Errorf("failed to start the http api server :%w", err)
	}

	return nil
}

func startHTTPSServer(handlers *handler.Handlers, x509Source *workloadapi.X509Source, tconfigdSpiffeID func() ([]spiffeid.ID, error), traTGenAuthorizedSpiffeIDs func() ([]spiffeid.ID, error), logger *zap.Logger) error {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/jwks.json", handlers.GetJwksHandler).Methods("GET")
	router.Handle("/token_endpoint", middlewares.AuthorizeSpiffeID(traTGenAuthorizedSpiffeIDs)(http.HandlerFunc(handlers.TokenEndpointHandler))).Methods("POST")
	router.Handle("/generation-endpoint-rule-webhook", middlewares.AuthorizeSpiffeID(tconfigdSpiffeID)(http.HandlerFunc(handlers.GenerationEndpointRuleWebhookHandler))).Methods("POST")
	router.Handle("/generation-token-rule-webhook", middlewares.AuthorizeSpiffeID(tconfigdSpiffeID)(http.HandlerFunc(handlers.GenerationTokenRuleWebhookHandler))).Methods("POST")

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", HTTPS_PORT),
		TLSConfig:    tlsconfig.MTLSServerConfig(x509Source, x509Source, tlsconfig.AuthorizeAny()),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	logger.Info("Starting HTTPS server...", zap.Int("port", HTTPS_PORT))

	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		logger.Error("Failed to start the https api server", zap.Error(err))

		return fmt.Errorf("failed to start the https api server :%w", err)
	}

	return nil
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()
}
