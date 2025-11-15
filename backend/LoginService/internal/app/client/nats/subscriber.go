package nats

import (
	"context"
	"encoding/json"
	"log"

	methods "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/nats-io/nats.go"
)

type NatsConn struct {
	js   nats.JetStreamContext
	core methods.CoreLogic
}

type DeactivateUserCommand struct {
	UserID string `json:"user_id"`
	Status bool   `json:"user_status"`
}

func NewNatsConn(js nats.JetStreamContext, core methods.CoreLogic) *NatsConn {
	return &NatsConn{js: js, core: core}
}

func (s *NatsConn) Start(ctx context.Context) {
	if _, err := s.js.Subscribe(eventbus.LoginDeactivateUserCommandTopic, func(msg *nats.Msg) {
		s.handleChangeUserStatus(msg)
	}, nats.ManualAck(), nats.AckExplicit(), nats.Durable("login-deactivate")); err != nil {
		log.Printf("NATS: Failed to subscribe to: %v", err)
		return
	}

	log.Printf("NATS: Subscribed to %s", eventbus.LoginDeactivateUserCommandTopic)

	<-ctx.Done()
}

func (s *NatsConn) handleChangeUserStatus(msg *nats.Msg) {
	log.Printf("NATS RECEIVE: Subject=%s", msg.Subject)

	var command DeactivateUserCommand
	if err := json.Unmarshal(msg.Data, &command); err != nil {
		log.Printf("NATS ERROR: Failed to unmarshal DeactivateUserCommand: %v", err)
		acknowledge(msg, true)
		return
	}

	if command.UserID == "" {
		log.Printf("NATS ERROR: Received DeactivateUserCommand with empty user_id")
		acknowledge(msg, true)
		return
	}

	err := s.core.ChangeUserStatus(context.Background(), command.UserID, command.Status)
	if err != nil {
		log.Printf("NATS ERROR: Failed to process deactivate command for user %s: %v", command.UserID, err)
		acknowledge(msg, false)
		return
	}

	log.Printf("NATS: Successfully processed deactivate command for user %s", command.UserID)
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
