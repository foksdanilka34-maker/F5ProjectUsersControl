package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/nats-io/nats.go"
)

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{
		nc: nc,
	}
}

func (p *Publisher) PublishTaskEvent(ctx context.Context, topic string, event any) error {
	if p.nc == nil {
		log.Printf("NATS connection is nil, skipping event publication")
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal task event: %w", err)
	}

	err = p.nc.Publish(eventbus.ProjectTasksTopic, data)
	if err != nil {
		return fmt.Errorf("failed to publish task event: %w", err)
	}

	log.Println("published task event")
	return nil
}

func (p *Publisher) PublishProjectEvent(ctx context.Context, topic string, event any) error {
	if p.nc == nil {
		log.Printf("NATS connection is nil, skipping event publication")
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal project event: %w", err)
	}

	err = p.nc.Publish(eventbus.ProjectProjectsTopic, data)
	if err != nil {
		return fmt.Errorf("failed to publish project event: %w", err)
	}

	log.Println("published project event")
	return nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}
