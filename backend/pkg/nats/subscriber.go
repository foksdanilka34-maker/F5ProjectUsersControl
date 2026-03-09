package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type EmployeeEventHandler func(ctx context.Context, event *EmployeeEvent) error
type LogEntryHandler func(ctx context.Context, entry *LogEntry) error

type Subscriber struct {
	js jetstream.JetStream
}

func NewSubscriber(nc *nats.Conn) (*Subscriber, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

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

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "logs",
		Subjects:  []string{"logs.*"},
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logs stream: %w", err)
	}

	return &Subscriber{js: js}, nil
}

func (s *Subscriber) SubscribeEmployeeEvents(ctx context.Context, handler EmployeeEventHandler) error {
	stream, err := s.js.Stream(ctx, "employee-events")
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "business-employee-sync",
		FilterSubject: "employee.event.*",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("SubscribeEmployeeEvents: context cancelled, stopping")
				return
			default:
			}

			msgs, err := cons.Fetch(10, jetstream.FetchMaxWait(5*time.Second))
			if err != nil {
				select {
				case <-time.After(2 * time.Second):
				case <-ctx.Done():
					return
				}
				continue
			}
			for msg := range msgs.Messages() {
				var event EmployeeEvent
				if err := json.Unmarshal(msg.Data(), &event); err != nil {
					msg.Nak()
					continue
				}
				if err := handler(ctx, &event); err != nil {
					msg.Nak()
					continue
				}
				msg.Ack()
			}
		}
	}()

	return nil
}

func (s *Subscriber) SubscribeLogs(ctx context.Context, handler LogEntryHandler) error {
	stream, err := s.js.Stream(ctx, "logs")
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "logs-processor",
		FilterSubject: "logs.*",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("SubscribeLogs: context cancelled, stopping")
				return
			default:
			}

			msgs, err := cons.Fetch(100, jetstream.FetchMaxWait(5*time.Second))
			if err != nil {
				select {
				case <-time.After(2 * time.Second):
				case <-ctx.Done():
					return
				}
				continue
			}
			for msg := range msgs.Messages() {
				var entry LogEntry
				if err := json.Unmarshal(msg.Data(), &entry); err != nil {
					msg.Nak()
					continue
				}
				if err := handler(ctx, &entry); err != nil {
					msg.Nak()
					continue
				}
				msg.Ack()
			}
		}
	}()

	return nil
}
