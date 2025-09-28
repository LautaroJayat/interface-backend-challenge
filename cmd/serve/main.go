package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"

	natsAdapter "messaging-app/internal/adapters/nats"
	"messaging-app/internal/adapters/postgres"
	"messaging-app/internal/application"
	"messaging-app/internal/ports"
)

func main() {
	// Load configuration
	fullConfig, err := application.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logLevel := slog.LevelInfo
	switch fullConfig.Logging.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	slogLogger := slog.New(handler)
	appLogger := ports.NewSlogAdapter(slogLogger)

	// Initialize database
	db, err := initializeDatabase(fullConfig, appLogger)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize NATS
	natsConn, err := initializeNATS(fullConfig, appLogger)
	if err != nil {
		log.Fatalf("Failed to initialize NATS: %v", err)
	}
	defer natsConn.Close()

	// Initialize adapters
	messageRepo := postgres.NewPostgreSQLMessageRepository(db, appLogger)
	publisher := natsAdapter.NewNATSMessagePublisher(natsConn, appLogger)

	// Create application with interfaces and HTTP configuration
	app := application.NewApplication(
		fullConfig.GetApplicationConfig(),
		appLogger,
		messageRepo,
		publisher,
		fullConfig.GetHTTPConfig(),
	)

	// Initialize and start application
	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	os.Exit(0)
}

func initializeDatabase(config application.FullConfig, logger ports.Logger) (*sql.DB, error) {
	dbConfig := postgres.Config{
		Host:            config.Database.Host,
		Port:            config.Database.Port,
		User:            config.Database.User,
		Password:        config.Database.Password,
		Database:        config.Database.Database,
		SSLMode:         config.Database.SSLMode,
		MaxConnections:  config.Database.MaxConnections,
		MaxIdleTime:     config.Database.MaxIdleTime,
		ConnMaxLifetime: config.Database.ConnMaxLifetime,
	}

	db, err := postgres.NewConnection(dbConfig, logger)
	if err != nil {
		return nil, err
	}


	return db, nil
}

func initializeNATS(config application.FullConfig, logger ports.Logger) (*nats.Conn, error) {
	natsConfig := natsAdapter.Config{
		URL:             config.NATS.URL,
		MaxReconnects:   config.NATS.MaxReconnects,
		ReconnectWait:   config.NATS.ReconnectWait,
		ConnectTimeout:  config.NATS.ConnectTimeout,
		RequestTimeout:  config.NATS.RequestTimeout,
		EnableJetStream: config.NATS.EnableJetStream,
		ClusterName:     config.NATS.ClusterName,
	}

	return natsAdapter.NewConnection(natsConfig, logger)
}