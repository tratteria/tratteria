package database

import (
	"database/sql"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func InitializeDB(logger *zap.Logger) (*sql.DB, error) {
	dbPath, exists := os.LookupEnv("STOCKS_DB_PATH")
	if !exists {
		panic("STOCKS_DB_PATH environment variable is not set")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open stocks database: %w", err)
	}

	logger.Info("Applying database migrations for stocks...")

	err = applyMigrations(db, "./migrations")
	if err != nil && err != migrate.ErrNoChange {
		logger.Fatal("Failed to apply migrations to stocks database.", zap.Error(err))
	} else if err == migrate.ErrNoChange {
		logger.Info("No new migrations to apply to stocks database.")
	} else {
		logger.Info("Migrations applied successfully to stocks database.")
	}

	logger.Info("Stocks database setup completed successfully.")

	return db, nil
}

func applyMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("could not create database driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3",
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance: %v", err)
	}

	return m.Up()
}
