package analytics

import (
	"context"
	"log"
	"time"

	repo "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
)

type MetricsWorker struct {
	storage        repo.ProjectStorage
	updateInterval time.Duration
	stopChan       chan struct{}
}

func NewMetricsWorker(storage repo.ProjectStorage, updateInterval time.Duration) *MetricsWorker {
	if updateInterval == 0 {
		updateInterval = 5 * time.Minute
	}
	return &MetricsWorker{
		storage:        storage,
		updateInterval: updateInterval,
		stopChan:       make(chan struct{}),
	}
}

func (w *MetricsWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.updateInterval)
	defer ticker.Stop()

	log.Printf("MetricsWorker started with interval: %v", w.updateInterval)

	for {
		select {
		case <-ticker.C:
			w.updateMetrics(ctx)
		case <-w.stopChan:
			log.Println("MetricsWorker stopped")
			return
		case <-ctx.Done():
			log.Println("MetricsWorker context cancelled")
			return
		}
	}
}

func (w *MetricsWorker) Stop() {
	close(w.stopChan)
}

func (w *MetricsWorker) updateMetrics(ctx context.Context) {
	log.Println("MetricsWorker: updating project metrics...")

	// В реальной системе здесь бы был запрос на получение списка всех проектов
	// и обновление метрик для каждого. Для простоты пока оставим логику как есть.
	// Метрики будут вычисляться on-demand при запросе GetProjectMetrics

	log.Println("MetricsWorker: metrics update cycle completed")
}
