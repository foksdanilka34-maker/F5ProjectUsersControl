package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	js nats.JetStreamContext
}

func NewPublisher(js nats.JetStreamContext) *Publisher {
	return &Publisher{js: js}
}

func (p *Publisher) PublishTaskEvent(ctx context.Context, topic string, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal task event: %w", err)
	}

	if err := p.publish(topic, data); err != nil {
		return err
	}

	log.Printf("published task event %s", topic)
	return nil
}

func (p *Publisher) PublishProjectEvent(ctx context.Context, topic string, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal project event: %w", err)
	}

	if err := p.publish(topic, data); err != nil {
		return err
	}

	log.Printf("published project event %s", topic)
	return nil
}

func (p *Publisher) Close() {
	// JetStream publisher does not own the connection lifecycle.
}

func (p *Publisher) publish(subject string, payload []byte) error {
	if p == nil || p.js == nil {
		return fmt.Errorf("jetstream context is not initialized for %s", subject)
	}

	if _, err := p.js.Publish(subject, payload); err != nil {
		log.Printf("NATS: error publishing %s: %v", subject, err)
		return err
	}

	return nil
}
