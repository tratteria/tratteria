package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/handler"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/keys"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/middleware"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/service"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/subjecttokenhandler"
)

type App struct {
	Router               *mux.Router
	Config               *config.AppConfig
	SpireJwtSource       *workloadapi.JWTSource
	SubjectTokenHandlers *subjecttokenhandler.TokenHandlers
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

	appConfig := config.GetAppConfig()

	err = keys.Initialize(appConfig)
	if err != nil {
		logger.Fatal("Error initializing keys:", zap.Error(err))
	}

	spireJwtSource, err := config.GetSpireJwtSource(appConfig.Spiffe.EndpointSocket)
	if err != nil {
		logger.Fatal("Unable to create SPIRE JWTSource for fetching JWT-SVIDs.", zap.Error(err))
	}

	logger.Info("Successfully created SPIRE JWTSource for fetching JWT-SVIDs.")

	defer spireJwtSource.Close()

	subjectTokenHandlers := subjecttokenhandler.GetTokenHandlers(appConfig.ClientAuthenticationMethods, logger)

	app := &App{
		Router:               mux.NewRouter(),
		Config:               appConfig,
		SpireJwtSource:       spireJwtSource,
		SubjectTokenHandlers: subjectTokenHandlers,
		Logger:               logger,
	}

	middleware := middleware.GetMiddleware(app.Config, app.SpireJwtSource, app.Logger)

	app.Router.Use(middleware)

	appService := service.NewService(app.Config, app.SpireJwtSource, app.SubjectTokenHandlers, app.Logger)
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
