package main

import (
	"context"
	"fmt"
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
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/logging"
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

	if err := logging.InitLogger(); err != nil {
		panic(err)
	}
	defer logging.Sync()

	mainLogger := logging.GetLogger("main")

	x509SrcCtx, cancel := context.WithTimeout(context.Background(), SPIRE_SOURCE_TIMEOUT)

	defer cancel()

	x509Source, err := workloadapi.NewX509Source(x509SrcCtx)
	if err != nil {
		mainLogger.Fatal("Failed to create SPIRE X.509 source", zap.Error(err))
	}

	defer x509Source.Close()

	appConfig, err := config.GetAppConfig()
	if err != nil {
		mainLogger.Fatal("Error getting application configuration.", zap.Error(err))
	}

	err = keys.Initialize()
	if err != nil {
		mainLogger.Fatal("Error initializing keys:", zap.Error(err))
	}

	httpClient := &http.Client{}
	generationRules := v1alpha1.NewGenerationRulesImp(httpClient)

	configSyncClient := configsync.NewClient(appConfig.TconfigdHost, appConfig.TconfigdSpiffeID, appConfig.MyNamespace, generationRules, x509Source, logging.GetLogger("config-sync"))

	go func() {
		if err := configSyncClient.Start(ctx); err != nil {
			mainLogger.Fatal("Config sync client stopped with error", zap.Error(err))
		}
	}()

	apiLogger := logging.GetLogger("api-server")
	apiService := service.NewService(generationRules, apiLogger)
	apiHandler := handler.NewHandlers(apiService, apiLogger)

	go func() {
		err := startHTTPServer(apiHandler, mainLogger)
		if err != nil {
			mainLogger.Fatal("HTTP server exited with error", zap.Error(err))
		}
	}()

	traTGenAuthorizedSpiffeIDs := func() ([]spiffeid.ID, error) { return generationRules.GetTokenGenerationAuthorizedServiceIds() }

	go func() {
		if err := startHTTPSServer(
			apiHandler,
			x509Source,
			traTGenAuthorizedSpiffeIDs,
			mainLogger,
		); err != nil {
			mainLogger.Fatal("HTTPS server exited with error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	mainLogger.Info("Shutting down tratteria...")
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

func startHTTPSServer(handlers *handler.Handlers, x509Source *workloadapi.X509Source, traTGenAuthorizedSpiffeIDs func() ([]spiffeid.ID, error), logger *zap.Logger) error {
	router := mux.NewRouter()

	router.Handle("/token_endpoint", middlewares.AuthorizeSpiffeID(traTGenAuthorizedSpiffeIDs)(http.HandlerFunc(handlers.TokenEndpointHandler))).Methods("POST")

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
