package main

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/SGNL-ai/Txn-Tokens-Demonstration-Services/handlers/stockhandler"
	"github.com/SGNL-ai/Txn-Tokens-Demonstration-Services/migrations"
	"github.com/SGNL-ai/Txn-Tokens-Demonstration-Services/pkg/stockservice"
	"github.com/golang-migrate/migrate/v4"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
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

	defer logger.Sync()

	// TODO: Place the sqlite database at appropriate location
	db, err := sql.Open("sqlite3", "./stock.db")
	if err != nil {
		logger.Fatal("Cannot open database.", zap.Error(err))
	}

	defer db.Close()

	logger.Info("Applying database migrations...")

	migrationPath := filepath.Join("migrations", "scripts", "stock")

	err = migrations.ApplyMigrations(db, migrationPath)
	if err != nil && err != migrate.ErrNoChange {
		logger.Fatal("Failed to apply migrations.", zap.Error(err))
	} else if err == migrate.ErrNoChange {
		logger.Info("No new migrations to apply.")
	} else {
		logger.Info("Migrations applied successfully.")
	}

	app := &App{
		Router: mux.NewRouter(),
		DB:     db,
		Logger: logger,
	}

	stockService := stockservice.NewService(db, app.Logger)
	stockHandlers := stockhandler.NewHandlers(stockService, app.Logger)

	app.initializeRoutes(stockHandlers)

	srv := &http.Server{
		Handler:      app.Router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on 8080.")
	log.Fatal(srv.ListenAndServe())
}

func (a *App) initializeRoutes(handlers *stockhandler.Handlers) {
	a.Router.HandleFunc("/stocks/search", handlers.SearchStocks).Methods("GET")
	a.Router.HandleFunc("/stocks/{id}", handlers.GetStockDetails).Methods("GET")
}
