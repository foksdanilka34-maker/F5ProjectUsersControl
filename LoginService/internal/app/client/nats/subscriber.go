package nats

import (
	"context"
	"encoding/json"

	methods "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
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
	_, err := s.n.Subscribe(DeactivateUserCommandSubject, s.handleChangeUserStatus)
	if err != nil {
		app.Logger.Error("NATS: Failed to subscribe to", zap.Error(err))
	}
	app.Logger.Info("NATS: Subscribed to", zap.String("dUser", DeactivateUserCommandSubject))

	select {}
}

func (s *NatsConn) handleChangeUserStatus(msg *nats.Msg) {
	app.Logger.Info("NATS RECEIVE: Subject=", zap.String("Subject", msg.Subject))

	var command DeactivateUserCommand
	if err := json.Unmarshal(msg.Data, &command); err != nil {
		app.Logger.Error("NATS ERROR: Failed to unmarshal DeactivateUserCommand:", zap.Error(err))
		return
	}

	if command.UserID == "" {
		app.Logger.Error("NATS ERROR: Received DeactivateUserCommand with empty user_id")
		return
	}

	err := s.core.ChangeUserStatus(context.Background(), command.UserID, command.Status)
	if err != nil {
		app.Logger.Error("NATS ERROR: Failed to process deactivate command for user", zap.String("deactiv", command.UserID), zap.Error(err))
	} else {
		app.Logger.Info("NATS: Successfully processed deactivate command for user", zap.String("user", command.UserID))
	}
}
