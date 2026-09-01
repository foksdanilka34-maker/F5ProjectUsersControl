package outbox

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
)

type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, body []byte) error
}

type EventBroadcaster interface {
	Broadcast(eventType string, payload any)
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(r *repo.RepositoryRegistry) error) error
}

type PollerConfig struct {
	PollInterval time.Duration
	BatchSize    int
	WorkerCount  int
}

type Poller struct {
	txManager   TxManager
	publisher   EventPublisher
	broadcaster EventBroadcaster
	cfg         PollerConfig
}

func NewPoller(txManager TxManager, publisher EventPublisher, broadcaster EventBroadcaster, cfg PollerConfig) *Poller {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 500 * time.Millisecond
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 25
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	return &Poller{
		txManager:   txManager,
		publisher:   publisher,
		broadcaster: broadcaster,
		cfg:         cfg,
	}
}

func (p *Poller) Start(ctx context.Context) {
	log.Printf("[Business Outbox Poller] Starting poller with %d workers, batch size %d",
		p.cfg.WorkerCount, p.cfg.BatchSize)

	tasksChan := make(chan dto.OutboxRecord, p.cfg.BatchSize*2)
	var wg sync.WaitGroup

	for i := 0; i < p.cfg.WorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case record, ok := <-tasksChan:
					if !ok {
						return
					}
					p.processRecord(ctx, record)
				}
			}
		}(i + 1)
	}

	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[Business Outbox Poller] Context cancelled, shutting down...")
			close(tasksChan)
			wg.Wait()
			return
		case <-ticker.C:
			p.pollBatch(ctx, tasksChan)
		}
	}
}

func (p *Poller) pollBatch(ctx context.Context, tasksChan chan<- dto.OutboxRecord) {
	var records []dto.OutboxRecord

	err := p.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		var err error
		records, err = r.Outbox().FetchPendingBatch(ctx, p.cfg.BatchSize)
		return err
	})

	if err != nil {
		if ctx.Err() == nil {
			log.Printf("[Business Outbox Poller] Error fetching batch: %v", err)
		}
		return
	}

	for _, rec := range records {
		select {
		case <-ctx.Done():
			return
		case tasksChan <- rec:
		}
	}
}

func (p *Poller) processRecord(ctx context.Context, record dto.OutboxRecord) {
	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var err error
	if p.publisher != nil {
		err = p.publisher.Publish(pubCtx, record.EventType, record.Payload)
	}

	_ = p.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		if err != nil {
			log.Printf("[Business Outbox Poller] Failed to publish %s (%s): %v", record.ID, record.EventType, err)
			return r.Outbox().MarkFailed(ctx, record.ID, err.Error())
		}
		return r.Outbox().MarkPublished(ctx, record.ID)
	})
}
