package nats

// import (
// 	// "context"
// 	// "encoding/json"
// 	// "fmt"

// 	// // "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/storage"
// 	// "github.com/nats-io/nats.go"
// 	// "go.uber.org/zap"
// )

// type Publisher struct {
// 	conn *nats.Conn
// }

// type DeactivareUserCommand struct {
// 	UserID string `json:"user_id"`
// 	Status bool   `json:"user_status"`
// }

// func NewPublisher(conn *nats.Conn) *Publisher {
// 	return &Publisher{
// 		conn: conn,
// 	}
// }

// // func (p *Publisher) PublishDeactivateUserCommand(ctx context.Context, userID string, status bool) error {
// // 	cmd := DeactivareUserCommand{
// // 		UserID: userID,
// // 		Status: status,
// // 	}

// // 	payload, err := json.Marshal(cmd)
// // 	if err != nil {
// // 		return fmt.Errorf("fail in marshal deactiv command %v", err)
// // 	}

// // 	msg := &nats.Msg{
// // 		Subject: DeactivateUserCommandSubject,
// // 		Data:    payload,
// // 	}

// // 	if err := p.conn.PublishMsg(msg); err != nil {
// // 		app.Logger.Error("NATS: error publishing deactivate command", zap.Error(err))
// // 		return fmt.Errorf("error publishing deactivate command %v", err)
// // 	}

// // 	return nil
// // }
