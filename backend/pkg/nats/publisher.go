package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	TopicEmployeeCreated = "employee.event.created"
	TopicEmployeeUpdated = "employee.event.updated"
	TopicEmployeeDeleted = "employee.event.deleted"
	TopicLogEntry        = "logs.entry"
)

type EmployeeEvent struct {
	UserID   int64   `json:"user_id"`
	FullName string  `json:"full_name"`
	PhotoURL *string `json:"photo_url,omitempty"`
}

type LogEntry struct {
	Service   string `json:"service"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data,omitempty"`
}

type Publisher struct {
	js jetstream.JetStream
}

func NewPublisher(nc *nats.Conn) (*Publisher, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// Ensure employee-events stream exists
	ctx := context.Background()
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "employee-events",
		Subjects:  []string{"employee.event.*"},
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create employee-events stream: %w", err)
	}

	return &Publisher{js: js}, nil
}

func (p *Publisher) publish(ctx context.Context, subject string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(ctx, subject, payload)
	return err
}

func (p *Publisher) PublishEmployeeCreated(ctx context.Context, userID int64, fullName string, photoURL *string) error {
	event := &EmployeeEvent{
		UserID:   userID,
		FullName: fullName,
		PhotoURL: photoURL,
	}
	return p.publish(ctx, TopicEmployeeCreated, event)
}

func (p *Publisher) PublishEmployeeUpdated(ctx context.Context, userID int64, fullName string, photoURL *string) error {
	event := &EmployeeEvent{
		UserID:   userID,
		FullName: fullName,
		PhotoURL: photoURL,
	}
	return p.publish(ctx, TopicEmployeeUpdated, event)
}

func (p *Publisher) PublishEmployeeDeleted(ctx context.Context, userID int64) error {
	event := &EmployeeEvent{UserID: userID}
	return p.publish(ctx, TopicEmployeeDeleted, event)
}

func (p *Publisher) PublishLog(ctx context.Context, entry *LogEntry) error {
	return p.publish(ctx, TopicLogEntry, entry)
}
