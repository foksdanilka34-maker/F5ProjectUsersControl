package events

import (
	"context"
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

func (p *Publisher) PublishTaskEvent(ctx context.Context, event *TaskEvent) error {
	if p.nc == nil {
		log.Printf("NATS connection is nil, skipping event publication")
		return nil
	}

	data, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal task event: %w", err)
	}

	err = p.nc.Publish(eventbus.ProjectTasksTopic, data)
	if err != nil {
		return fmt.Errorf("failed to publish task event: %w", err)
	}

	log.Printf("published task event: type=%s, taskID=%s", event.EventType, event.TaskID)
	return nil
}

func (p *Publisher) PublishProjectEvent(ctx context.Context, event *ProjectEvent) error {
	if p.nc == nil {
		log.Printf("NATS connection is nil, skipping event publication")
		return nil
	}

	data, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal project event: %w", err)
	}

	err = p.nc.Publish(eventbus.ProjectProjectsTopic, data)
	if err != nil {
		return fmt.Errorf("failed to publish project event: %w", err)
	}

	log.Printf("published project event: type=%s, projectID=%s", event.EventType, event.ProjectID)
	return nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}
