package application

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	httpAdapter "messaging-app/internal/adapters/http"
)

type FullConfig struct {
	Server struct {
		Port         int           `mapstructure:"port"`
		Host         string        `mapstructure:"host"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	} `mapstructure:"server"`

	Auth struct {
		UserIDHeader  string `mapstructure:"user_id_header"`
		EmailHeader   string `mapstructure:"email_header"`
		HandlerHeader string `mapstructure:"handler_header"`
	} `mapstructure:"auth"`

	CORS struct {
		AllowedOrigins []string `mapstructure:"allowed_origins"`
		AllowedMethods []string `mapstructure:"allowed_methods"`
		AllowedHeaders []string `mapstructure:"allowed_headers"`
	} `mapstructure:"cors"`

	Database struct {
		Host            string        `mapstructure:"host"`
		Port            int           `mapstructure:"port"`
		User            string        `mapstructure:"user"`
		Password        string        `mapstructure:"password"`
		Database        string        `mapstructure:"database"`
		SSLMode         string        `mapstructure:"ssl_mode"`
		MaxConnections  int           `mapstructure:"max_connections"`
		MaxIdleTime     time.Duration `mapstructure:"max_idle_time"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	} `mapstructure:"database"`

	NATS struct {
		URL             string        `mapstructure:"url"`
		MaxReconnects   int           `mapstructure:"max_reconnects"`
		ReconnectWait   time.Duration `mapstructure:"reconnect_wait"`
		ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
		RequestTimeout  time.Duration `mapstructure:"request_timeout"`
		EnableJetStream bool          `mapstructure:"enable_jetstream"`
		ClusterName     string        `mapstructure:"cluster_name"`
	} `mapstructure:"nats"`

	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`

	Environment string `mapstructure:"environment"`
}

func LoadConfig() (FullConfig, error) {
	var config FullConfig

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")

	viper.SetDefault("auth.user_id_header", "x-interface-user-id")
	viper.SetDefault("auth.email_header", "x-interface-user-email")
	viper.SetDefault("auth.handler_header", "x-interface-user-handler")

	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{})

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.database", "messaging_app")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.max_idle_time", "15m")
	viper.SetDefault("database.conn_max_lifetime", "1h")

	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.max_reconnects", 10)
	viper.SetDefault("nats.reconnect_wait", "2s")
	viper.SetDefault("nats.connect_timeout", "5s")
	viper.SetDefault("nats.request_timeout", "10s")
	viper.SetDefault("nats.enable_jetstream", false)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("environment", "development")

	// Read from environment variables
	viper.SetEnvPrefix("MESSAGING_APP")
	viper.AutomaticEnv()

	// Read from config file if it exists
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults and environment variables
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// GetApplicationConfig extracts only the application-level config
func (fc FullConfig) GetApplicationConfig() Config {
	return Config{
		Server:      fc.Server,
		Environment: fc.Environment,
	}
}

// GetHTTPConfig extracts HTTP server configuration
func (fc FullConfig) GetHTTPConfig() httpAdapter.Config {
	return httpAdapter.Config{
		Host:         fc.Server.Host,
		Port:         fc.Server.Port,
		ReadTimeout:  fc.Server.ReadTimeout,
		WriteTimeout: fc.Server.WriteTimeout,
		Auth: httpAdapter.AuthConfig{
			UserIDHeader:  fc.Auth.UserIDHeader,
			EmailHeader:   fc.Auth.EmailHeader,
			HandlerHeader: fc.Auth.HandlerHeader,
		},
		CORS: httpAdapter.CORSConfig{
			AllowedOrigins: fc.CORS.AllowedOrigins,
			AllowedMethods: fc.CORS.AllowedMethods,
			AllowedHeaders: fc.CORS.AllowedHeaders,
		},
	}
}