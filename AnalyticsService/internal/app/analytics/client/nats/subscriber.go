package nats

import (
	"context"
	"encoding/json"
	"time"

	//"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"log"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn *nats.Conn
	core *core.Core
}

func NewSubscriber(conn *nats.Conn, core *core.Core) *Subscriber {
	return &Subscriber{
		conn: conn,
		core: core,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	_, err := s.conn.Subscribe(eventbus.EventTypeTaskAssigned, s.handleAssignTask)
	if err != nil {
		return err
	}

	log.Println("NATS subscriber started")
	log.Printf("Subscribed to topics: %s", eventbus.EventTypeTaskAssigned)

	select {}
}

func (s *Subscriber) handleAssignTask(msg *nats.Msg) {
	log.Println("Started handleAssignTask")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	taskEvent := &eventbus.TaskEvent{}
	err := json.Unmarshal(msg.Data, taskEvent)
	if err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}
	if taskEvent.AssigneeID == nil {
		log.Printf("AssigneeID is nil, skipping")
		return
	}
	metrics := &analytics.EmployeeMetrics{
		EmployeeID:      *taskEvent.AssigneeID,
		InProgressTasks: 1,
	}
	data, err := s.core.GetEmployeeMetrics(ctx, *taskEvent.AssigneeID)
	if err != nil {
		now := time.Now()
		metrics.MetricDate = now
		metrics.AssignedTasks = 1
	} else {
		metrics.AssignedTasks = data.AssignedTasks + 1
		metrics.InProgressTasks = data.InProgressTasks + 1
		metrics.CompletedTasks = data.CompletedTasks
		metrics.OverdueTasks = data.OverdueTasks
		metrics.OnTimeCompletionTask = data.OnTimeCompletionTask
		metrics.TotalPrioritWeight = data.TotalPrioritWeight
		metrics.TotalTaskDurationSeconds = data.TotalTaskDurationSeconds
		metrics.CreatedAt = data.CreatedAt
		metrics.ID = data.ID
		metrics.MetricDate = data.MetricDate
	}
	err = s.core.SaveEmployeeMetrics(ctx, metrics)
	if err != nil {
		log.Printf("NATS ERROR saving employee :%v", err)
		return
	}
}

func (s *Subscriber) handleTaskCompleted(msg *nats.Msg) {
	log.Println("Start handelTaskCompleted")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	taskEvent := &eventbus.TaskEvent{}
	err := json.Unmarshal(msg.Data, taskEvent)
	if err != nil {
		log.Printf("error unmarshaling data %v", err)
		return
	}
	metrics, err := s.core.GetEmployeeMetrics(ctx, *taskEvent.AssigneeID)
	if err != nil {
		log.Printf("error getting Employee metrics %v", err)
	}
	metrics.AssignedTasks -= 1
	metrics.InProgressTasks -= 1
	metrics.CompletedTasks += 1
	err = s.core.SaveEmployeeMetrics(ctx, metrics)
	if err != nil {
		log.Printf("error handling tasks complete %v", err)
	}
}



// func (s *Subscriber) handleTaskCompleted(ctx context.Context, msg *nats.Msg) {
// 	var event TaskCompletedEvent
// 	if err := json.Unmarshal(msg.Data, &event); err != nil {
// 		log.Printf("failed to unmarshal task completed event: %v", err)
// 		return
// 	}

// 	metrics, err := s.core.GetEmployeeMetrics(ctx, event.EmployeeID, time.Now().AddDate(0, 0, -1), time.Now())
// 	if err != nil {
// 		log.Printf("failed to get employee metrics: %v", err)
// 		return
// 	}

// 	if len(metrics) == 0 {
// 		metrics = append(metrics, &analytics.EmployeeMetrics{
// 			EmployeeID: event.EmployeeID,
// 			MetricDate: time.Now(),
// 		})
// 	}

// 	metrics[0].TasksCompleted++

// 	onTimeRate := calculateOnTimeRate(event.Duration, 8.0)
// 	if metrics[0].OnTimeCompletionRate == 0 {
// 		metrics[0].OnTimeCompletionRate = onTimeRate
// 	} else {
// 		metrics[0].OnTimeCompletionRate = (metrics[0].OnTimeCompletionRate + onTimeRate) / 2
// 	}

// 	metrics[0].EfficiencyScore = calculateEfficiencyScore(metrics[0])

// 	if err := s.core.SaveEmployeeMetrics(ctx, metrics[0]); err != nil {
// 		log.Printf("failed to save employee metrics: %v", err)
// 	}
// }
