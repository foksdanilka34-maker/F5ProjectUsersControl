package storage

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsConfig struct {
	URL string
}

type NATSClient struct {
	Conn *nats.Conn
	JS   nats.JetStreamContext
}

func NewNATSConnection(cfg NatsConfig) (*NATSClient, error) {
	natsConnection, err := nats.Connect(cfg.URL, nats.Timeout(time.Second*5))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	jetStream, err := natsConnection.JetStream()
	if err != nil {
		natsConnection.Close()
		return nil, fmt.Errorf("failed to initialize JetStream context: %w", err)
	}

	return &NATSClient{
		Conn: natsConnection,
		JS:   jetStream,
	}, nil
}

func (c *NATSClient) Close() {
	if c == nil || c.Conn == nil {
		return
	}

	_ = c.Conn.Drain()
	c.Conn.Close()
}


