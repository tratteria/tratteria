package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/handlers"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/database"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/service"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
	Logger *zap.Logger
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
		logger.Fatal("Stocks database initialization failed.", zap.Error(err))
	}

	defer db.Close()

	app := &App{
		Router: mux.NewRouter(),
		DB:     db,
		Logger: logger,
	}

	stockService := service.NewService(db, app.Logger)
	stockHandlers := handler.NewHandlers(stockService, app.Logger)

	app.initializeRoutes(stockHandlers)

	srv := &http.Server{
		Handler:      app.Router,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on 8080.")
	log.Fatal(srv.ListenAndServe())
}

func (a *App) initializeRoutes(handlers *handler.Handlers) {
	a.Router.HandleFunc("/stocks/search", handlers.SearchStocks).Methods("GET")
	a.Router.HandleFunc("/stocks/{id}", handlers.GetStockDetails).Methods("GET")
}
