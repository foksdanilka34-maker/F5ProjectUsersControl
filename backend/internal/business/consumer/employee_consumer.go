package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EmployeeEventPayload struct {
	EventID   string  `json:"event_id"`
	UserID    int64   `json:"user_id"`
	FullName  string  `json:"full_name"`
	PhotoURL  *string `json:"photo_url,omitempty"`
	Role      string  `json:"role"`
	Timestamp int64   `json:"timestamp"`
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(r *repo.RepositoryRegistry) error) error
}

type EmployeeConsumer struct {
	rabbit    *rabbitmq.Client
	txManager TxManager
}

func NewEmployeeConsumer(rabbit *rabbitmq.Client, txManager TxManager) *EmployeeConsumer {
	return &EmployeeConsumer{
		rabbit:    rabbit,
		txManager: txManager,
	}
}

func (c *EmployeeConsumer) Start(ctx context.Context, concurrency int) error {
	log.Printf("[RabbitMQ Consumer] Starting consumer pool (%d workers) on queue %s...",
		concurrency, rabbitmq.EmployeeSyncQueue)

	return c.rabbit.Consume(ctx, rabbitmq.EmployeeSyncQueue, concurrency, c.handleMessage)
}

func (c *EmployeeConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	var payload EmployeeEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		log.Printf("[Consumer] Fatal: failed to unmarshal message body: %v", err)
		return err
	}

	if payload.EventID == "" || payload.UserID == 0 {
		log.Printf("[Consumer] Invalid event payload: event_id=%s, user_id=%d", payload.EventID, payload.UserID)
		return fmt.Errorf("invalid event payload")
	}

	return c.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		alreadyProcessed, err := r.Inbox().IsProcessed(ctx, payload.EventID)
		if err != nil {
			return fmt.Errorf("failed to check idempotency: %w", err)
		}
		if alreadyProcessed {
			log.Printf("[Consumer] Event %s already processed, skipping", payload.EventID)
			return nil
		}

		switch msg.RoutingKey {
		case rabbitmq.EmployeeCreatedKey, rabbitmq.EmployeeUpdatedKey:
			if err := r.Inbox().UpsertUserMeta(ctx, payload.UserID, payload.FullName, payload.PhotoURL); err != nil {
				return fmt.Errorf("failed to upsert user meta: %w", err)
			}
			log.Printf("[Consumer] Synchronized user_meta for user %d (%s)", payload.UserID, payload.FullName)

		case rabbitmq.EmployeeDeletedKey:
			if err := r.Inbox().DeleteUserMeta(ctx, payload.UserID); err != nil {
				return fmt.Errorf("failed to delete user meta: %w", err)
			}
			log.Printf("[Consumer] Deleted user_meta for user %d", payload.UserID)

		default:
			log.Printf("[Consumer] Unknown routing key %s, recording event", msg.RoutingKey)
		}

		return r.Inbox().MarkProcessed(ctx, payload.EventID, msg.RoutingKey)
	})
}
