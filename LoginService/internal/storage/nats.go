package storage

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsConfig struct {
	URL string
}

func NewNATSConnection(cfg NatsConfig) (*nats.Conn, error) {
	natsConnetion, err := nats.Connect(cfg.URL, nats.Timeout(time.Second*5))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return natsConnetion, nil
}
