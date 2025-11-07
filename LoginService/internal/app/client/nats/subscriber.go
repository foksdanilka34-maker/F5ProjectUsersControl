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
	n    *nats.Conn
	core methods.CoreLogic
}

func NewNatsConn(ns *nats.Conn, core methods.CoreLogic) *NatsConn {
	return &NatsConn{
		n:    ns,
		core: core,
	}
}

type DeactivateUserCommand struct {
	UserID string `json:"user_id"`
	Status bool   `json:"user_status"`
}

func (s *NatsConn) Start() {
	_, err := s.n.Subscribe(eventbus.LoginDeactivateUserCommandTopic, s.handleChangeUserStatus)
	if err != nil {
		log.Printf("NATS: Failed to subscribe to: %v", err)
	}
	log.Printf("NATS: Subscribed to %s", eventbus.LoginDeactivateUserCommandTopic)

	select {}
}

func (s *NatsConn) handleChangeUserStatus(msg *nats.Msg) {
	log.Printf("NATS RECEIVE: Subject=%s", msg.Subject)

	var command DeactivateUserCommand
	if err := json.Unmarshal(msg.Data, &command); err != nil {
		log.Printf("NATS ERROR: Failed to unmarshal DeactivateUserCommand: %v", err)
		return
	}

	if command.UserID == "" {
		log.Printf("NATS ERROR: Received DeactivateUserCommand with empty user_id")
		return
	}

	err := s.core.ChangeUserStatus(context.Background(), command.UserID, command.Status)
	if err != nil {
		log.Printf("NATS ERROR: Failed to process deactivate command for user %s: %v", command.UserID, err)
	} else {
		log.Printf("NATS: Successfully processed deactivate command for user %s", command.UserID)
	}
}
