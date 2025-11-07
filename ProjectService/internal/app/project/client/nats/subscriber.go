package nats

import (
	"context"
	"encoding/json"
	"log"

	repo "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn    *nats.Conn
	storage repo.UserMetaStorage
}

type EmployeeEvent struct {
	UserID   string  `json:"user_id"`
	FullName string  `json:"full_name"`
	PhotoURL *string `json:"photo_url,omitempty"`
}

func NewSubscriber(conn *nats.Conn, storage repo.UserMetaStorage) *Subscriber {
	return &Subscriber{
		conn:    conn,
		storage: storage,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	_, err := s.conn.Subscribe(eventbus.EmployeeCreatedEventTopic, func(msg *nats.Msg) {
		s.handleEmployeeCreated(ctx, msg)
	})
	if err != nil {
		return err
	}
	log.Printf("NATS: subscribed to %s", eventbus.EmployeeCreatedEventTopic)

	_, err = s.conn.Subscribe(eventbus.EmployeeUpdatedEventTopic, func(msg *nats.Msg) {
		s.handleEmployeeUpdated(ctx, msg)
	})
	if err != nil {
		return err
	}
	log.Printf("NATS: subscribed to %s", eventbus.EmployeeUpdatedEventTopic)

	_, err = s.conn.Subscribe(eventbus.EmployeeDeletedEventTopic, func(msg *nats.Msg) {
		s.handleEmployeeDeleted(ctx, msg)
	})
	if err != nil {
		return err
	}
	log.Printf("NATS: subscribed to %s", eventbus.EmployeeDeletedEventTopic)

	return nil
}

func (s *Subscriber) handleEmployeeCreated(ctx context.Context, msg *nats.Msg) {
	var event EmployeeEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("NATS: failed to unmarshal employee created event: %v", err)
		return
	}

	log.Printf("NATS: received employee created event: userID=%s, name=%s", event.UserID, event.FullName)

	photoURL := ""
	if event.PhotoURL != nil {
		photoURL = *event.PhotoURL
	}

	if err := s.storage.UpsertUserMeta(ctx, event.UserID, event.FullName, photoURL); err != nil {
		log.Printf("NATS: failed to upsert user meta: %v", err)
		return
	}

	log.Printf("NATS: user meta created/updated successfully: userID=%s", event.UserID)
}

func (s *Subscriber) handleEmployeeUpdated(ctx context.Context, msg *nats.Msg) {
	var event EmployeeEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("NATS: failed to unmarshal employee updated event: %v", err)
		return
	}

	log.Printf("NATS: received employee updated event: userID=%s, name=%s", event.UserID, event.FullName)

	photoURL := ""
	if event.PhotoURL != nil {
		photoURL = *event.PhotoURL
	}

	if err := s.storage.UpsertUserMeta(ctx, event.UserID, event.FullName, photoURL); err != nil {
		log.Printf("NATS: failed to update user meta: %v", err)
		return
	}

	log.Printf("NATS: user meta updated successfully: userID=%s", event.UserID)
}

func (s *Subscriber) handleEmployeeDeleted(ctx context.Context, msg *nats.Msg) {
	var event EmployeeEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("NATS: failed to unmarshal employee deleted event: %v", err)
		return
	}

	log.Printf("NATS: received employee deleted event: userID=%s", event.UserID)

	if err := s.storage.DeleteUserMeta(ctx, event.UserID); err != nil {
		log.Printf("NATS: failed to delete user meta: %v", err)
		return
	}

	log.Printf("NATS: user meta deleted successfully: userID=%s", event.UserID)
}
