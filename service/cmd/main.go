package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/tratteria/tratteria/handler"
	"github.com/tratteria/tratteria/pkg/accessevaluation"
	"github.com/tratteria/tratteria/pkg/config"
	"github.com/tratteria/tratteria/pkg/keys"
	"github.com/tratteria/tratteria/pkg/middleware"
	"github.com/tratteria/tratteria/pkg/service"
	"github.com/tratteria/tratteria/pkg/spire"
	"github.com/tratteria/tratteria/pkg/subjecttokenhandler"
)

type App struct {
	Router               *mux.Router
	Config               *config.AppConfig
	SpireJwtSource       *workloadapi.JWTSource
	SubjectTokenHandlers *subjecttokenhandler.TokenHandlers
	HttpClient           *http.Client
	AccessEvaluator      accessevaluation.AccessEvaluatorService
	Logger               *zap.Logger
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

	if len(os.Args) < 2 {
		logger.Error("No configuration file provided. Please specify the configuration path as an argument when running the service.",
			zap.String("usage", "tratteria <config-path>"))
		os.Exit(1)
	}

	configPath := os.Args[1]

	appConfig := config.GetAppConfig(configPath)

	err = keys.Initialize(appConfig)
	if err != nil {
		logger.Fatal("Error initializing keys:", zap.Error(err))
	}

	httpClient := &http.Client{}

	var accessEvaluator accessevaluation.AccessEvaluatorService
	if appConfig.EnableAccessEvaluation {
		accessEvaluator = accessevaluation.NewAccessEvaluator(appConfig.AccessEvaluationAPI, httpClient)
	} else {
		accessEvaluator = &accessevaluation.NoOpAccessEvaluator{}
	}

	spireJwtSource, err := spire.GetSpireJwtSource(appConfig, logger)
	if err != nil {
		logger.Fatal("Unable to create SPIRE JWTSource for fetching JWT-SVIDs.", zap.Error(err))
	}

	if spireJwtSource != nil {
		logger.Info("Successfully created SPIRE JWTSource for fetching JWT-SVIDs.")

		defer spireJwtSource.Close()
	}

	subjectTokenHandlers := subjecttokenhandler.GetTokenHandlers(appConfig.SubjectTokens, logger)

	app := &App{
		Router:               mux.NewRouter(),
		Config:               appConfig,
		SpireJwtSource:       spireJwtSource,
		SubjectTokenHandlers: subjectTokenHandlers,
		HttpClient:           httpClient,
		AccessEvaluator:      accessEvaluator,
		Logger:               logger,
	}

	middleware := middleware.GetMiddleware(app.Config, app.SpireJwtSource, app.Logger)

	app.Router.Use(middleware)

	appService := service.NewService(app.Config, app.SpireJwtSource, app.SubjectTokenHandlers, app.AccessEvaluator, app.Logger)
	appHandler := handler.NewHandlers(appService, app.Config, app.Logger)

	app.initializeRoutes(appHandler)

	srv := &http.Server{
		Handler:      app.Router,
		Addr:         "0.0.0.0:9090",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on 9090.")
	log.Fatal(srv.ListenAndServe())
}

func (a *App) initializeRoutes(handlers *handler.Handlers) {
	a.Router.HandleFunc("/.well-known/jwks.json", handlers.GetJwksHandler).Methods("GET")
	a.Router.HandleFunc("/token_endpoint", handlers.TokenEndpointHandler).Methods("POST")
}
