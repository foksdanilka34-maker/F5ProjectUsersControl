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
	js nats.JetStreamContext
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

func NewPublisher(js nats.JetStreamContext) *Publisher {
	return &Publisher{
		js: js,
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

	if err := p.publish(eventbus.LoginDeactivateUserCommandTopic, payload); err != nil {
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

	if err := p.publish(eventbus.EmployeeCreatedEventTopic, payload); err != nil {
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

	if err := p.publish(eventbus.EmployeeUpdatedEventTopic, payload); err != nil {
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

	if err := p.publish(eventbus.EmployeeDeletedEventTopic, payload); err != nil {
		return fmt.Errorf("error publishing employee deleted event: %v", err)
	}

	log.Printf("NATS: employee deleted event published: userID=%s", userID)
	return nil
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
