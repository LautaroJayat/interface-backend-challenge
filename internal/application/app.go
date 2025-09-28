package application

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "messaging-app/internal/adapters/http"
	httphandlers "messaging-app/internal/handlers/http"
	"messaging-app/internal/ports"
)

type Application struct {
	config     Config
	logger     ports.Logger
	httpServer *httpAdapter.Server
}

type Config struct {
	Server struct {
		Port         int           `mapstructure:"port"`
		Host         string        `mapstructure:"host"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	} `mapstructure:"server"`

	Environment string `mapstructure:"environment"`
}

func NewApplication(
	config Config,
	logger ports.Logger,
	messageRepo ports.MessageRepository,
	publisher ports.MessagePublisher,
	httpConfig httpAdapter.Config,
) *Application {
	// Create HTTP server adapter with full configuration
	httpServer := httpAdapter.NewServer(httpConfig, logger)

	// Initialize route providers
	messageRoutes := httphandlers.NewMessageRoutes(messageRepo, publisher, logger)
	chatRoutes := httphandlers.NewChatRoutes(messageRepo, logger)

	// Collect all routes
	var allRoutes []httpAdapter.Route
	allRoutes = append(allRoutes, messageRoutes.GetRoutes()...)
	allRoutes = append(allRoutes, chatRoutes.GetRoutes()...)

	// Register routes with the server
	httpServer.RegisterRoutes(allRoutes)

	return &Application{
		config:     config,
		logger:     logger,
		httpServer: httpServer,
	}
}

func (app *Application) Initialize() error {
	app.logger.Info("Initializing application...")

	// Initialize HTTP server
	if err := app.httpServer.Initialize(); err != nil {
		return err
	}

	app.logger.Info("Application initialized successfully")
	return nil
}

func (app *Application) Start() error {
	app.logger.Info("Starting application...",
		"port", app.config.Server.Port,
		"environment", app.config.Environment,
	)

	// Start HTTP server in goroutine
	go func() {
		if err := app.httpServer.Start(); err != nil {
			app.logger.Error("HTTP server failed", "error", err)
		}
	}()

	app.logger.Info("Application started successfully",
		"address", app.httpServer.Address(),
	)

	// Wait for interrupt signal
	return app.waitForShutdown()
}

func (app *Application) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	app.logger.Info("Shutting down application...")

	return app.Shutdown()
}

func (app *Application) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := app.httpServer.Shutdown(ctx); err != nil {
		app.logger.Error("Failed to shutdown HTTP server", "error", err)
	}

	app.logger.Info("Application shutdown completed")
	return nil
}