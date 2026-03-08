package server

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/logs/core"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	LogsStreamName  = "LOGS"
	LogsSubjectName = "logs.>"
)

type LogMessage struct {
	Service   string            `json:"service"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	UserID    string            `json:"user_id,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

type LogsServer struct {
	logService *core.LogService
	nc         *nats.Conn
	js         jetstream.JetStream
}

func NewLogsServer(logService *core.LogService, nc *nats.Conn) (*LogsServer, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return &LogsServer{
		logService: logService,
		nc:         nc,
		js:         js,
	}, nil
}

func (s *LogsServer) Start(ctx context.Context) error {

	stream, err := s.getOrCreateStream(ctx)
	if err != nil {
		return err
	}

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "logs-processor",
		FilterSubject: "logs.*",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	log.Println("Logs server started, consuming messages...")

	go s.consumeMessages(ctx, cons)

	return nil
}

func (s *LogsServer) getOrCreateStream(ctx context.Context) (jetstream.Stream, error) {
	stream, err := s.js.Stream(ctx, LogsStreamName)
	if err == nil {
		return stream, nil
	}

	stream, err = s.js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      LogsStreamName,
		Subjects:  []string{LogsSubjectName},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour, // 7 days retention
	})
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (s *LogsServer) consumeMessages(ctx context.Context, cons jetstream.Consumer) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgs, err := cons.Fetch(100, jetstream.FetchMaxWait(time.Second*5))
			if err != nil {
				continue
			}

			for msg := range msgs.Messages() {
				if err := s.processMessage(ctx, msg); err != nil {
					log.Printf("Failed to process log message: %v", err)
					msg.Nak()
				} else {
					msg.Ack()
				}
			}
		}
	}
}

func (s *LogsServer) processMessage(ctx context.Context, msg jetstream.Msg) error {
	var logMsg LogMessage
	if err := json.Unmarshal(msg.Data(), &logMsg); err != nil {
		return err
	}

	req := &core.CreateLogRequest{
		Service:  logMsg.Service,
		Level:    logMsg.Level,
		Message:  logMsg.Message,
		Metadata: logMsg.Metadata,
	}

	if logMsg.UserID != "" {
		req.UserID = &logMsg.UserID
	}
	if logMsg.RequestID != "" {
		req.RequestID = &logMsg.RequestID
	}

	return s.logService.Log(ctx, req)
}


