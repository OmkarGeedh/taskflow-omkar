// Package database manages the PostgreSQL connection pool.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Global database instance
var DB *sql.DB

// InitDatabase initialises the PostgreSQL connection pool using the given URL.
func InitDatabase(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verify reachability
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db

	zap.L().Info("database connection established",
		zap.String("driver", "pgx"),
		zap.Int("max_open_conns", 50),
		zap.Int("max_idle_conns", 10),
		zap.Duration("conn_max_lifetime", 5*time.Minute),
	)

	return nil
}

// GetDB returns the global database instance.
func GetDB() *sql.DB {
	return DB
}

// CloseDatabase closes the database connection.
func CloseDatabase() error {
	if DB == nil {
		return nil
	}
	return DB.Close()
}

// HealthCheck performs a database health check.
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialised")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}
