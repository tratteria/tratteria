package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/handler"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/database"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/service"
)

type App struct {
	Router     *mux.Router
	DB         *sql.DB
	Logger     *zap.Logger
	HTTPClient *http.Client
	Config     *config.OrderConfig
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

	cfg := config.NewConfig()

	app := &App{
		Router:     mux.NewRouter(),
		DB:         db,
		Logger:     logger,
		HTTPClient: &http.Client{},
		Config:     cfg,
	}

	orderService := service.NewService(db, app.Logger, app.HTTPClient, app.Config)
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
