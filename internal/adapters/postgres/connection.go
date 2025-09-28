package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"messaging-app/internal/ports"

	_ "github.com/lib/pq"
)

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConnections  int
	MaxIdleTime     time.Duration
	ConnMaxLifetime time.Duration
}

func NewConnection(config Config, logger ports.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxConnections / 2)
	db.SetConnMaxIdleTime(config.MaxIdleTime)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database,
		"max_connections", config.MaxConnections,
	)

	return db, nil
}

func DefaultConfig() Config {
	return Config{
		Host:            "127.0.0.1",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "messaging_app",
		SSLMode:         "disable",
		MaxConnections:  25,
		MaxIdleTime:     15 * time.Minute,
		ConnMaxLifetime: time.Hour,
	}
}
