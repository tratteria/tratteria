package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/handler"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/configsync"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/generationrules/v1alpha1"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/keys"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/middleware"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/service"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/spire"
)

const (
	ServerPort = 9090
)

type App struct {
	Router          *mux.Router
	Config          *config.AppConfig
	SpireJwtSource  *workloadapi.JWTSource
	HttpClient      *http.Client
	GenerationRules *v1alpha1.GenerationRulesImp
	Logger          *zap.Logger
}

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

	spireJwtSource, err := spire.GetSpireJwtSource()
	if err != nil {
		logger.Fatal("Unable to create SPIRE JWTSource for fetching JWT-SVIDs.", zap.Error(err))
	}

	if spireJwtSource != nil {
		logger.Info("Successfully created SPIRE JWTSource for fetching JWT-SVIDs.")

		defer spireJwtSource.Close()
	}

	configSyncClient, err := configsync.NewClient(ServerPort, appConfig.TconfigdUrl, appConfig.MyNamespace, generationRules, httpClient, logger)
	if err != nil {
		logger.Fatal("Error creating configuration sync client for tconfigd", zap.Error(err))
	}

	if err := configSyncClient.Start(); err != nil {
		logger.Fatal("Error establishing communication with tconfigd", zap.Error(err))
	}

	app := &App{
		Router:          mux.NewRouter(),
		Config:          appConfig,
		SpireJwtSource:  spireJwtSource,
		HttpClient:      httpClient,
		GenerationRules: generationRules,
		Logger:          logger,
	}

	middleware := middleware.GetMiddleware(app.Config, app.GenerationRules, app.SpireJwtSource, app.Logger)

	app.Router.Use(middleware)

	appService := service.NewService(app.SpireJwtSource, app.GenerationRules, app.Logger)
	appHandler := handler.NewHandlers(appService, app.Logger)

	app.initializeRoutes(appHandler)

	srv := &http.Server{
		Handler:      app.Router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", ServerPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on 9090.")
	log.Fatal(srv.ListenAndServe())
}

func (a *App) initializeRoutes(handlers *handler.Handlers) {
	a.Router.HandleFunc("/token_endpoint", handlers.TokenEndpointHandler).Methods("POST")
	a.Router.HandleFunc("/.well-known/jwks.json", handlers.GetJwksHandler).Methods("GET")

	a.Router.HandleFunc("/generation-rules", handlers.GetGenerationRulesHandler).Methods("GET")
	a.Router.HandleFunc("/generation-endpoint-rule-webhook", handlers.GenerationEndpointRuleWebhookHandler).Methods("POST")
	a.Router.HandleFunc("/generation-token-rule-webhook", handlers.GenerationTokenRuleWebhookHandler).Methods("POST")
}
