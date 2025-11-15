package nats

import (
	"context"
	"encoding/json"
	"log"

	projectRepo "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	js      nats.JetStreamContext
	storage projectRepo.UserMetaStorage
}

type EmployeeEvent struct {
	UserID   string  `json:"user_id"`
	FullName string  `json:"full_name"`
	PhotoURL *string `json:"photo_url,omitempty"`
}

func NewSubscriber(js nats.JetStreamContext, storage projectRepo.UserMetaStorage) *Subscriber {
	return &Subscriber{
		js:      js,
		storage: storage,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	subscriptions := []struct {
		subject string
		handler func(context.Context, *nats.Msg)
		durable string
	}{
		{eventbus.EmployeeCreatedEventTopic, s.handleEmployeeCreated, "project-employee-created"},
		{eventbus.EmployeeUpdatedEventTopic, s.handleEmployeeUpdated, "project-employee-updated"},
		{eventbus.EmployeeDeletedEventTopic, s.handleEmployeeDeleted, "project-employee-deleted"},
	}

	for _, sub := range subscriptions {
		if _, err := s.js.Subscribe(sub.subject, func(msg *nats.Msg) {
			sub.handler(ctx, msg)
		}, nats.ManualAck(), nats.AckExplicit(), nats.Durable(sub.durable)); err != nil {
			return err
		}
		log.Printf("NATS: subscribed to %s", sub.subject)
	}

	<-ctx.Done()
	return ctx.Err()
}

func (s *Subscriber) handleEmployeeCreated(ctx context.Context, msg *nats.Msg) {
	var event EmployeeEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("NATS: failed to unmarshal employee created event: %v", err)
		acknowledge(msg, true)
		return
	}

	log.Printf("NATS: received employee created event: userID=%s, name=%s", event.UserID, event.FullName)

	photoURL := ""
	if event.PhotoURL != nil {
		photoURL = *event.PhotoURL
	}

	if err := s.storage.UpsertUserMeta(ctx, event.UserID, event.FullName, photoURL); err != nil {
		log.Printf("NATS: failed to upsert user meta: %v", err)
		acknowledge(msg, false)
		return
	}

	log.Printf("NATS: user meta created/updated successfully: userID=%s", event.UserID)
	acknowledge(msg, true)
}

func (s *Subscriber) handleEmployeeUpdated(ctx context.Context, msg *nats.Msg) {
	var event EmployeeEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("NATS: failed to unmarshal employee updated event: %v", err)
		acknowledge(msg, true)
		return
	}

	log.Printf("NATS: received employee updated event: userID=%s, name=%s", event.UserID, event.FullName)

	photoURL := ""
	if event.PhotoURL != nil {
		photoURL = *event.PhotoURL
	}

	if err := s.storage.UpsertUserMeta(ctx, event.UserID, event.FullName, photoURL); err != nil {
		log.Printf("NATS: failed to update user meta: %v", err)
		acknowledge(msg, false)
		return
	}

	log.Printf("NATS: user meta updated successfully: userID=%s", event.UserID)
	acknowledge(msg, true)
}

func (s *Subscriber) handleEmployeeDeleted(ctx context.Context, msg *nats.Msg) {
	var event EmployeeEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("NATS: failed to unmarshal employee deleted event: %v", err)
		acknowledge(msg, true)
		return
	}

	log.Printf("NATS: received employee deleted event: userID=%s", event.UserID)

	if err := s.storage.DeleteUserMeta(ctx, event.UserID); err != nil {
		log.Printf("NATS: failed to delete user meta: %v", err)
		acknowledge(msg, false)
		return
	}

	log.Printf("NATS: user meta deleted successfully: userID=%s", event.UserID)
	acknowledge(msg, true)
}

func acknowledge(msg *nats.Msg, success bool) {
	if msg == nil {
		return
	}
	if success {
		if err := msg.Ack(); err != nil {
			log.Printf("NATS: failed to ack message %s: %v", msg.Subject, err)
		}
		return
	}
	if err := msg.Nak(); err != nil {
		log.Printf("NATS: failed to nak message %s: %v", msg.Subject, err)
	}
}
