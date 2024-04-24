package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/handler"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/database"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/middleware"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/service"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/trats"
)

type App struct {
	Router         *mux.Router
	DB             *sql.DB
	HTTPClient     *http.Client
	Config         *config.OrderConfig
	SpireJwtSource *workloadapi.JWTSource
	TraTsVerifer   *trats.Verifier
	Logger         *zap.Logger
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

	db, err := database.InitializeDB(logger)
	if err != nil {
		logger.Fatal("Order database initialization failed.", zap.Error(err))
	}

	defer db.Close()

	orderConfig := config.GetOrderConfig()

	spireJwtSource := config.GetSpireJwtSource(logger)

	defer spireJwtSource.Close()

	traTsVerifier := config.GetTraTsVerifier()

	app := &App{
		Router:         mux.NewRouter(),
		DB:             db,
		HTTPClient:     &http.Client{},
		Config:         orderConfig,
		SpireJwtSource: spireJwtSource,
		TraTsVerifer:   traTsVerifier,
		Logger:         logger,
	}

	middleware := middleware.GetMiddleware(app.Config, app.SpireJwtSource, app.TraTsVerifer, app.Logger)

	app.Router.Use(middleware)

	orderService := service.NewService(app.DB, app.HTTPClient, app.Config, app.SpireJwtSource, app.Logger)
	orderHandler := handler.NewHandlers(orderService, app.Logger)

	app.initializeRoutes(orderHandler)

	srv := &http.Server{
		Handler:      app.Router,
		Addr:         "0.0.0.0:8090",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on 8090.")
	log.Fatal(srv.ListenAndServe())
}

func (a *App) initializeRoutes(handlers *handler.Handlers) {
	a.Router.HandleFunc("/api/order", handlers.OrderHandler).Methods("POST")
	a.Router.HandleFunc("/api/order/{id}", handlers.GetOrderDetailsHandler).Methods("GET")
}
