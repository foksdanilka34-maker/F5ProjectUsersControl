package nats

import (
	"time"
	"context"
	//"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/core"
	projectpb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	//"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn          *nats.Conn
	core          *core.Core
	projectClient projectpb.ProjectServiceClient
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

type ProjectTaskEvent struct {
	EventType   string     `json:"event_type"`
	TaskID      string     `json:"task_id"`
	ProjectID   string     `json:"project_id"`
	Status      string     `json:"status"`
	OldStatus   string     `json:"old_status,omitempty"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	CreatorID   string     `json:"creator_id,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}

func NewSubscriber(conn *nats.Conn, core *core.Core, projectClient projectpb.ProjectServiceClient) *Subscriber {
	return &Subscriber{
		conn:          conn,
		core:          core,
		projectClient: projectClient,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	return nil
}

// 	_, err = s.conn.Subscribe(eventbus.EmployeeTaskAssignedTopic, func(msg *nats.Msg) {
// 		s.handleTaskAssigned(ctx, msg)
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	_, err = s.conn.Subscribe(eventbus.ProjectTasksTopic, func(msg *nats.Msg) {
// 		s.handleProjectStatusChanged(ctx, msg)
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	log.Println("NATS subscriber started")
// 	log.Printf("Subscribed to topics: %s, %s, %s",
// 		eventbus.EmployeeTaskCompletedTopic,
// 		eventbus.EmployeeTaskAssignedTopic,
// 		eventbus.ProjectTasksTopic)
// 	return nil
// }

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

// func (s *Subscriber) handleTaskAssigned(ctx context.Context, msg *nats.Msg) {
// 	var event TaskAssignedEvent
// 	if err := json.Unmarshal(msg.Data, &event); err != nil {
// 		log.Printf("failed to unmarshal task assigned event: %v", err)
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

// 	metrics[0].TasksAssigned++
// 	metrics[0].EfficiencyScore = calculateEfficiencyScore(metrics[0])

// 	if err := s.core.SaveEmployeeMetrics(ctx, metrics[0]); err != nil {
// 		log.Printf("failed to save employee metrics: %v", err)
// 	}
// }

// func (s *Subscriber) handleProjectStatusChanged(ctx context.Context, msg *nats.Msg) {
// 	// Log RAW event data for debugging
// 	log.Printf("=== RAW NATS MESSAGE ===")
// 	log.Printf("Subject: %s", msg.Subject)
// 	log.Printf("Data: %s", string(msg.Data))
// 	log.Printf("========================")

// 	var event ProjectTaskEvent
// 	if err := json.Unmarshal(msg.Data, &event); err != nil {
// 		log.Printf("ERROR: failed to unmarshal project task event: %v", err)
// 		log.Printf("Raw data was: %s", string(msg.Data))
// 		return
// 	}

// 	log.Printf("✓ Parsed task event: type=%s, taskID=%s, projectID=%s, status=%s, oldStatus=%s, assigneeID=%v",
// 		event.EventType, event.TaskID, event.ProjectID, event.Status, event.OldStatus, event.AssigneeID)

// 	// Handle task assignment (when assignee is set)
// 	if event.AssigneeID != nil && *event.AssigneeID != "" {
// 		s.handleTaskAssignmentFromEvent(ctx, &event)
// 	}

// 	// Handle task completion (when status changed to DONE)
// 	if event.Status == "DONE" && event.OldStatus != "DONE" {
// 		s.handleTaskCompletionFromEvent(ctx, &event)
// 	}

// 	// Update project metrics
// 	s.updateProjectMetricsFromEvent(ctx, &event)
// }

// func (s *Subscriber) handleTaskAssignmentFromEvent(ctx context.Context, event *ProjectTaskEvent) {
// 	if event.AssigneeID == nil || *event.AssigneeID == "" {
// 		return
// 	}

// 	employeeID := *event.AssigneeID
// 	metrics, err := s.core.GetEmployeeMetrics(ctx, employeeID, time.Now().AddDate(0, 0, -1), time.Now())
// 	if err != nil {
// 		log.Printf("failed to get employee metrics for assignment: %v", err)
// 		return
// 	}

// 	if len(metrics) == 0 {
// 		metrics = append(metrics, &analytics.EmployeeMetrics{
// 			EmployeeID: employeeID,
// 			MetricDate: time.Now(),
// 		})
// 	}

// 	metrics[0].TasksAssigned++
// 	metrics[0].EfficiencyScore = calculateEfficiencyScore(metrics[0])

// 	if err := s.core.SaveEmployeeMetrics(ctx, metrics[0]); err != nil {
// 		log.Printf("failed to save employee metrics after assignment: %v", err)
// 	} else {
// 		log.Printf("saved employee metrics after assignment: employeeID=%s, tasksAssigned=%d",
// 			employeeID, metrics[0].TasksAssigned)
// 	}
// }

// func (s *Subscriber) handleTaskCompletionFromEvent(ctx context.Context, event *ProjectTaskEvent) {
// 	if event.AssigneeID == nil || *event.AssigneeID == "" {
// 		log.Printf("task completed but no assignee: taskID=%s", event.TaskID)
// 		return
// 	}

// 	employeeID := *event.AssigneeID
// 	metrics, err := s.core.GetEmployeeMetrics(ctx, employeeID, time.Now().AddDate(0, 0, -1), time.Now())
// 	if err != nil {
// 		log.Printf("failed to get employee metrics for completion: %v", err)
// 		return
// 	}

// 	if len(metrics) == 0 {
// 		metrics = append(metrics, &analytics.EmployeeMetrics{
// 			EmployeeID: employeeID,
// 			MetricDate: time.Now(),
// 		})
// 	}

// 	metrics[0].TasksCompleted++

// 	// Calculate on-time rate based on due date vs completion date
// 	var onTimeRate float64 = 100.0
// 	if event.DueDate != nil && event.CompletedAt != nil {
// 		if event.CompletedAt.After(*event.DueDate) {
// 			onTimeRate = 0.0
// 			metrics[0].OverdueTasks++
// 		}
// 	}

// 	if metrics[0].OnTimeCompletionRate == 0 {
// 		metrics[0].OnTimeCompletionRate = onTimeRate
// 	} else {
// 		metrics[0].OnTimeCompletionRate = (metrics[0].OnTimeCompletionRate + onTimeRate) / 2
// 	}

// 	metrics[0].EfficiencyScore = calculateEfficiencyScore(metrics[0])
// 	metrics[0].TaskCompletionRate = calculateCompletionRate(metrics[0])

// 	if err := s.core.SaveEmployeeMetrics(ctx, metrics[0]); err != nil {
// 		log.Printf("failed to save employee metrics after completion: %v", err)
// 	} else {
// 		log.Printf("saved employee metrics after completion: employeeID=%s, completed=%d, onTimeRate=%.2f",
// 			employeeID, metrics[0].TasksCompleted, metrics[0].OnTimeCompletionRate)
// 	}
// }

// func (s *Subscriber) updateProjectMetricsFromEvent(ctx context.Context, event *ProjectTaskEvent) {

// 	projectMetrics, err := s.core.GetProjectMetrics(ctx, event.ProjectID, time.Now().AddDate(0, 0, -1), time.Now())
// 	if err != nil {
// 		log.Printf("failed to get project metrics: %v", err)
// 		return
// 	}

// 	if len(projectMetrics) == 0 {
// 		// Get project details from ProjectService
// 		project, err := s.projectClient.GetProject(ctx, &projectpb.GetProjectRequest{
// 			ProjectId: event.ProjectID,
// 		})
// 		if err != nil {
// 			log.Printf("❌ Failed to get project details for %s: %v", event.ProjectID, err)
// 			return
// 		}

// 		projectMetrics = append(projectMetrics, &analytics.ProjectMetrics{
// 			ProjectID:   event.ProjectID,
// 			ProjectName: project.Name,
// 			ManagerID:   project.ManagerId,
// 			ManagerName: "", // TODO: Get manager name from EmployeeService
// 			MetricDate:  time.Now(),
// 		})
// 	}

// 	metrics := projectMetrics[0]

// 	if event.Status == "DONE" && event.OldStatus != "DONE" {
// 		metrics.CompletedTasks++
// 	}

// 	if event.Status == "IN_PROGRESS" && event.OldStatus == "TODO" {
// 		metrics.InProgressTasks++
// 	}

// 	if event.CompletedAt != nil && event.DueDate != nil {
// 		if event.CompletedAt.After(*event.DueDate) {
// 			metrics.OverdueTasks++
// 		}
// 	}

// 	if err := s.core.SaveProjectMetrics(ctx, metrics); err != nil {
// 		log.Printf("failed to save project metrics: %v", err)
// 	}
// }

// func calculateOnTimeRate(actualDuration, expectedDuration float64) float64 {
// 	if actualDuration <= expectedDuration {
// 		return 100.0
// 	}
// 	return (expectedDuration / actualDuration) * 100
// }

// func calculateCompletionRate(metrics *analytics.EmployeeMetrics) float64 {
// 	if metrics.TasksAssigned == 0 {
// 		return 0
// 	}
// 	return float64(metrics.TasksCompleted) / float64(metrics.TasksAssigned) * 100
// }

// func calculateEfficiencyScore(metrics *analytics.EmployeeMetrics) float64 {
// 	if metrics.TasksAssigned == 0 {
// 		return 0
// 	}

// 	completionRate := float64(metrics.TasksCompleted) / float64(metrics.TasksAssigned) * 100
// 	if completionRate > 100 {
// 		completionRate = 100
// 	}

// 	score := (completionRate + metrics.OnTimeCompletionRate) / 2
// 	if score > 100 {
// 		score = 100
// 	}

// 	return score
// }
