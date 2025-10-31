package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

type DeactivareUserCommand struct {
	UserID string `json:"user_id"`
	Status bool   `json:"user_status"`
}

type EmployeeEvent struct {
	UserID   string  `json:"user_id"`
	FullName string  `json:"full_name"`
	PhotoURL *string `json:"photo_url,omitempty"`
}

func NewPublisher(conn *nats.Conn) *Publisher {
	return &Publisher{
		conn: conn,
	}
}

func (p *Publisher) PublishDeactivateUserCommand(ctx context.Context, userID string, status bool) error {
	cmd := DeactivareUserCommand{
		UserID: userID,
		Status: status,
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("fail in marshal deactiv command %v", err)
	}

	msg := &nats.Msg{
		Subject: DeactivateUserCommandSubject,
		Data:    payload,
	}

	if err := p.conn.PublishMsg(msg); err != nil {
		log.Printf("NATS: error publishing deactivate command: %v", err)
		return fmt.Errorf("error publishing deactivate command %v", err)
	}

	return nil
}

func (p *Publisher) PublishEmployeeCreated(ctx context.Context, userID, fullName string, photoURL *string) error {
	event := EmployeeEvent{
		UserID:   userID,
		FullName: fullName,
		PhotoURL: photoURL,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal employee created event: %v", err)
	}

	msg := &nats.Msg{
		Subject: EmployeeCreatedEvent,
		Data:    payload,
	}

	if err := p.conn.PublishMsg(msg); err != nil {
		log.Printf("NATS: error publishing employee created event: %v", err)
		return fmt.Errorf("error publishing employee created event: %v", err)
	}

	log.Printf("NATS: employee created event published: userID=%s", userID)
	return nil
}

func (p *Publisher) PublishEmployeeUpdated(ctx context.Context, userID, fullName string, photoURL *string) error {
	event := EmployeeEvent{
		UserID:   userID,
		FullName: fullName,
		PhotoURL: photoURL,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal employee updated event: %v", err)
	}

	msg := &nats.Msg{
		Subject: EmployeeUpdatedEvent,
		Data:    payload,
	}

	if err := p.conn.PublishMsg(msg); err != nil {
		log.Printf("NATS: error publishing employee updated event: %v", err)
		return fmt.Errorf("error publishing employee updated event: %v", err)
	}

	log.Printf("NATS: employee updated event published: userID=%s", userID)
	return nil
}

func (p *Publisher) PublishEmployeeDeleted(ctx context.Context, userID string) error {
	event := EmployeeEvent{
		UserID: userID,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal employee deleted event: %v", err)
	}

	msg := &nats.Msg{
		Subject: EmployeeDeletedEvent,
		Data:    payload,
	}

	if err := p.conn.PublishMsg(msg); err != nil {
		log.Printf("NATS: error publishing employee deleted event: %v", err)
		return fmt.Errorf("error publishing employee deleted event: %v", err)
	}

	log.Printf("NATS: employee deleted event published: userID=%s", userID)
	return nil
}
