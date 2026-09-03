package consumer

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Sealer interface {
	Open(ciphertext []byte) (string, error)
}

type ExtensionConsumer struct {
	rabbit    *rabbitmq.Client
	txManager TxManager
	sealer    Sealer
	http      *http.Client
}

func NewExtensionConsumer(rabbit *rabbitmq.Client, txManager TxManager, sealer Sealer) *ExtensionConsumer {
	return &ExtensionConsumer{
		rabbit:    rabbit,
		txManager: txManager,
		sealer:    sealer,
		http:      &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *ExtensionConsumer) Start(ctx context.Context, concurrency int) error {
	log.Printf("[Extension Dispatch Consumer] Starting consumer pool (%d workers) on queue %s...",
		concurrency, rabbitmq.ExtensionDispatchQueue)

	return c.rabbit.Consume(ctx, rabbitmq.ExtensionDispatchQueue, concurrency, c.handleMessage)
}

func (c *ExtensionConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	category := eventCategory(msg.RoutingKey)
	if category == "" {
		return nil
	}

	var probe struct {
		ProjectID int64 `json:"project_id"`
	}
	if err := json.Unmarshal(msg.Body, &probe); err != nil || probe.ProjectID == 0 {
		return nil
	}

	var targets []dto.ExtensionDTO
	err := c.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		var err error
		targets, err = r.Extension().ListEnabledForProjectByEvent(ctx, probe.ProjectID, category)
		return err
	})
	if err != nil {
		return err
	}

	for _, ext := range targets {
		c.deliver(ctx, ext, category, msg.Body)
	}

	return nil
}

func (c *ExtensionConsumer) deliver(ctx context.Context, ext dto.ExtensionDTO, category string, payload []byte) {
	secret, err := c.sealer.Open(ext.SharedSecretEnc)
	if err != nil {
		log.Printf("[Extension Dispatch Consumer] Failed to decrypt secret for %s: %v", ext.Key, err)
		return
	}

	envelope, err := json.Marshal(map[string]any{
		"event": category,
		"data":  json.RawMessage(payload),
	})
	if err != nil {
		log.Printf("[Extension Dispatch Consumer] Failed to build envelope for %s: %v", ext.Key, err)
		return
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(envelope)
	signature := hex.EncodeToString(mac.Sum(nil))

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, ext.BaseURL+"/webhooks", bytes.NewReader(envelope))
	if err != nil {
		log.Printf("[Extension Dispatch Consumer] Failed to build request for %s: %v", ext.Key, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-F5-Event", category)
	req.Header.Set("X-F5-Signature", "sha256="+signature)

	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("[Extension Dispatch Consumer] Delivery to %s (%s) failed: %v", ext.Key, ext.BaseURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("[Extension Dispatch Consumer] Delivery to %s rejected with status %d", ext.Key, resp.StatusCode)
		return
	}

	log.Printf("[Extension Dispatch Consumer] Delivered %s to %s", category, ext.Key)
}

func eventCategory(routingKey string) string {
	switch routingKey {
	case "task.event.created":
		return dto.ExtEventTaskCreated
	case "task.event.moved", "task.event.review_requested", "task.event.reviewed", "task.event.approved":
		return dto.ExtEventTaskStatusChanged
	case "task.event.comment_added":
		return dto.ExtEventCommentAdded
	default:
		return ""
	}
}
