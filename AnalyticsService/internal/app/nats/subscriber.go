package nats

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/core"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn *nats.Conn
	core *core.Core
}

type TaskCompletedEvent struct {
	EmployeeID  string    `json:"employee_id"`
	TaskID      string    `json:"task_id"`
	ProjectID   string    `json:"project_id"`
	CompletedAt time.Time `json:"completed_at"`
	Priority    string    `json:"priority"`
	Duration    float64   `json:"duration_hours"`
}

type TaskAssignedEvent struct {
	EmployeeID string `json:"employee_id"`
	TaskID     string `json:"task_id"`
	ProjectID  string `json:"project_id"`
	Priority   string `json:"priority"`
}

type ProjectStatusChangedEvent struct {
	ProjectID string    `json:"project_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewSubscriber(conn *nats.Conn, core *core.Core) *Subscriber {
	return &Subscriber{
		conn: conn,
		core: core,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	_, err := s.conn.Subscribe("events.employee.task.completed", func(msg *nats.Msg) {
		s.handleTaskCompleted(ctx, msg)
	})
	if err != nil {
		return err
	}

	_, err = s.conn.Subscribe("events.employee.task.assigned", func(msg *nats.Msg) {
		s.handleTaskAssigned(ctx, msg)
	})
	if err != nil {
		return err
	}

	_, err = s.conn.Subscribe("events.project.task.status_changed", func(msg *nats.Msg) {
		s.handleProjectStatusChanged(ctx, msg)
	})
	if err != nil {
		return err
	}

	log.Println("NATS subscriber started")
	return nil
}

func (s *Subscriber) handleTaskCompleted(ctx context.Context, msg *nats.Msg) {
	var event TaskCompletedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("failed to unmarshal task completed event: %v", err)
		return
	}

	metrics, err := s.core.GetEmployeeMetrics(ctx, event.EmployeeID, time.Now().AddDate(0, 0, -1), time.Now())
	if err != nil {
		log.Printf("failed to get employee metrics: %v", err)
		return
	}

	if len(metrics) == 0 {
		metrics = append(metrics, &analytics.EmployeeMetrics{
			EmployeeID: event.EmployeeID,
			MetricDate: time.Now(),
		})
	}

	metrics[0].TasksCompleted++
	metrics[0].AvgCompletionTimeHours = (metrics[0].AvgCompletionTimeHours + event.Duration) / 2

	onTimeRate := calculateOnTimeRate(event.Duration, 8.0)
	if metrics[0].OnTimeCompletionRate == 0 {
		metrics[0].OnTimeCompletionRate = onTimeRate
	} else {
		metrics[0].OnTimeCompletionRate = (metrics[0].OnTimeCompletionRate + onTimeRate) / 2
	}

	metrics[0].EfficiencyScore = calculateEfficiencyScore(metrics[0])

	if err := s.core.SaveEmployeeMetrics(ctx, metrics[0]); err != nil {
		log.Printf("failed to save employee metrics: %v", err)
	}
}

func (s *Subscriber) handleTaskAssigned(ctx context.Context, msg *nats.Msg) {
	var event TaskAssignedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("failed to unmarshal task assigned event: %v", err)
		return
	}

	metrics, err := s.core.GetEmployeeMetrics(ctx, event.EmployeeID, time.Now().AddDate(0, 0, -1), time.Now())
	if err != nil {
		log.Printf("failed to get employee metrics: %v", err)
		return
	}

	if len(metrics) == 0 {
		metrics = append(metrics, &analytics.EmployeeMetrics{
			EmployeeID: event.EmployeeID,
			MetricDate: time.Now(),
		})
	}

	metrics[0].TasksAssigned++
	metrics[0].EfficiencyScore = calculateEfficiencyScore(metrics[0])

	if err := s.core.SaveEmployeeMetrics(ctx, metrics[0]); err != nil {
		log.Printf("failed to save employee metrics: %v", err)
	}
}

func (s *Subscriber) handleProjectStatusChanged(ctx context.Context, msg *nats.Msg) {
	var event ProjectStatusChangedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("failed to unmarshal project status changed event: %v", err)
		return
	}

	log.Printf("Project %s status changed to %s", event.ProjectID, event.Status)
}

func calculateOnTimeRate(actualDuration, expectedDuration float64) float64 {
	if actualDuration <= expectedDuration {
		return 100.0
	}
	return (expectedDuration / actualDuration) * 100
}

func calculateEfficiencyScore(metrics *analytics.EmployeeMetrics) float64 {
	if metrics.TasksAssigned == 0 {
		return 0
	}

	completionRate := float64(metrics.TasksCompleted) / float64(metrics.TasksAssigned) * 100
	if completionRate > 100 {
		completionRate = 100
	}

	score := (completionRate + metrics.OnTimeCompletionRate) / 2
	if score > 100 {
		score = 100
	}

	return score
}
