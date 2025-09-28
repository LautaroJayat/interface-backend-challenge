package nats

import (
	"fmt"
	"time"

	"messaging-app/internal/ports"

	"github.com/nats-io/nats.go"
)

type Config struct {
	URL             string
	MaxReconnects   int
	ReconnectWait   time.Duration
	ConnectTimeout  time.Duration
	RequestTimeout  time.Duration
	EnableJetStream bool
	ClusterName     string
}

func NewConnection(config Config, logger ports.Logger) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("messaging-app"),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
		nats.Timeout(config.ConnectTimeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Info("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info("NATS connection established",
		"url", config.URL,
		"server_info", conn.ConnectedServerName(),
		"cluster", conn.ConnectedClusterName(),
	)

	return conn, nil
}

func DefaultConfig() Config {
	return Config{
		URL:             nats.DefaultURL,
		MaxReconnects:   10,
		ReconnectWait:   2 * time.Second,
		ConnectTimeout:  5 * time.Second,
		RequestTimeout:  10 * time.Second,
		EnableJetStream: false,
		ClusterName:     "",
	}
}
