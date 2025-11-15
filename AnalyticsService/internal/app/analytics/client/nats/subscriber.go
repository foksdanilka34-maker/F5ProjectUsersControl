package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	js   nats.JetStreamContext
	core *core.Core
}

func NewSubscriber(js nats.JetStreamContext, core *core.Core) *Subscriber {
	return &Subscriber{js: js, core: core}
}

func (s *Subscriber) Start(ctx context.Context) error {
	subscriptions := []struct {
		subject string
		handler func(context.Context, *nats.Msg)
		durable string
	}{
		{eventbus.EventTypeTaskCreated, s.handleTaskCreated, "task-created"},
		{eventbus.EventTypeTaskAssigned, s.handleAssignTask, "task-assigned"},
		{eventbus.EventTypeTaskCompleted, s.handleTaskCompleted, "task-completed"},
		{eventbus.EventTypeTaskDeleted, s.handleTaskDeleted, "task-deleted"},
		{eventbus.EventTypeTaskStatusChanged, s.handleTaskStatusChanged, "task-status"},
		{eventbus.EventTypeProjectCreated, s.handleProjectCreated, "project-created"},
		{eventbus.EventTypeProjectUpdated, s.handleProjectUpdated, "project-updated"},
		{eventbus.EventTypeProjectMemberAdd, s.handleProjectMemberAdded, "project-member-add"},
		{eventbus.EventTypeProjectMemberDel, s.handleProjectMemberRemoved, "project-member-del"},
	}

	for _, sub := range subscriptions {
		if _, err := s.js.Subscribe(sub.subject, func(msg *nats.Msg) {
			sub.handler(ctx, msg)
		}, nats.ManualAck(), nats.AckExplicit(), nats.Durable(fmt.Sprintf("analytics-%s", sub.durable))); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", sub.subject, err)
		}
		log.Printf("NATS: subscribed to %s", sub.subject)
	}

	<-ctx.Done()
	return ctx.Err()
}

func (s *Subscriber) handleTaskCreated(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleTaskCreated")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	taskEvent := &eventbus.TaskEvent{}
	if err := json.Unmarshal(msg.Data, taskEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if taskEvent.AssigneeID == nil {
		log.Println("AssigneeID is nil, skipping")
		return
	}

	if err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
		m.AssignedTasks++
		m.InProgressTasks++
	}); err != nil {
		log.Printf("NATS ERROR updating employee metrics for new task: %v", err)
		success = false
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		p.TotalTasks++
	}); err != nil {
		log.Printf("NATS ERROR updating project metrics for new task: %v", err)
		success = false
		return
	}

	log.Println("Completed handleTaskCreated")
}

func (s *Subscriber) handleTaskDeleted(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleTaskDeleted")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	taskEvent := &eventbus.TaskEvent{}
	if err := json.Unmarshal(msg.Data, taskEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if taskEvent.AssigneeID == nil {
		log.Println("AssigneeID is nil, skipping")
		return
	}

	if err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
		if m.AssignedTasks > 0 {
			m.AssignedTasks--
		}
		if taskEvent.OldStatus == "IN_PROGRESS" {
			m.InProgressTasks--
		}
	}); err != nil {
		log.Printf("NATS ERROR updating employee metrics for deleted task: %v", err)
		success = false
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if p.TotalTasks > 0 {
			p.TotalTasks--
		}
	}); err != nil {
		log.Printf("NATS ERROR updating project metrics for deleted task: %v", err)
		success = false
		return
	}

	log.Println("Completed handleTaskDeleted")
}

func (s *Subscriber) handleTaskStatusChanged(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleTaskStatusChanged")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	taskEvent := &eventbus.TaskEvent{}
	if err := json.Unmarshal(msg.Data, taskEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if taskEvent.AssigneeID == nil {
		log.Println("AssigneeID is nil, skipping")
		return
	}

	if err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
		if taskEvent.OldStatus == "IN_PROGRESS" {
			m.InProgressTasks--
		}
		if taskEvent.Status == "IN_PROGRESS" {
			m.InProgressTasks++
		}
	}); err != nil {
		log.Printf("NATS ERROR updating employee metrics for status change: %v", err)
		success = false
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if taskEvent.OldStatus == "IN_PROGRESS" {
			p.InProgressTasks--
		}
		if taskEvent.Status == "IN_PROGRESS" {
			p.InProgressTasks++
		}
	}); err != nil {
		log.Printf("NATS ERROR updating project metrics for status change: %v", err)
		success = false
		return
	}

	log.Println("Completed handleTaskStatusChanged")
}

func (s *Subscriber) handleAssignTask(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleAssignTask")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	taskEvent := &eventbus.TaskEvent{}
	if err := json.Unmarshal(msg.Data, taskEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if taskEvent.PrevAssigneeID != nil {
		if err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.PrevAssigneeID, func(m *analytics.EmployeeMetrics) {
			m.AssignedTasks--
			if taskEvent.Status == "IN_PROGRESS" {
				m.InProgressTasks--
			}
		}); err != nil {
			log.Printf("NATS ERROR updating metrics for previous assignee: %v", err)
			success = false
			return
		}
	}

	if taskEvent.AssigneeID != nil {
		if err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
			m.AssignedTasks++
			if taskEvent.Status == "IN_PROGRESS" {
				m.InProgressTasks++
			}
		}); err != nil {
			log.Printf("NATS ERROR updating employee metrics for new assignee: %v", err)
			success = false
			return
		}
	}

	log.Println("Completed handleAssignTask")
}

func (s *Subscriber) handleTaskCompleted(ctx context.Context, msg *nats.Msg) {
	log.Println("Start handelTaskCompleted")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	taskEvent := &eventbus.TaskEvent{}
	if err := json.Unmarshal(msg.Data, taskEvent); err != nil {
		log.Printf("error unmarshaling data %v", err)
		return
	}

	if taskEvent.AssigneeID == nil {
		log.Printf("AssigneeID is nil, skipping")
		return
	}

	if err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
		if m.InProgressTasks > 0 {
			m.InProgressTasks--
		}
		m.CompletedTasks++

		if taskEvent.DueDate != nil && taskEvent.CompletedAt != nil {
			if taskEvent.CompletedAt.Before(*taskEvent.DueDate) || taskEvent.CompletedAt.Equal(*taskEvent.DueDate) {
				m.OnTimeCompletionTask++
			}
		}

		if taskEvent.StartedAt != nil && taskEvent.CompletedAt != nil {
			duration := taskEvent.CompletedAt.Sub(*taskEvent.StartedAt)
			m.TotalTaskDurationSeconds += int64(duration.Seconds())
		}
	}); err != nil {
		log.Printf("error handling tasks complete %v", err)
		success = false
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if p.InProgressTasks > 0 {
			p.InProgressTasks--
		}
		p.CompletedTasks++

		if taskEvent.DueDate != nil && taskEvent.CompletedAt != nil {
			if taskEvent.CompletedAt.Before(*taskEvent.DueDate) || taskEvent.CompletedAt.Equal(*taskEvent.DueDate) {
				p.OnTimeCompletedTasks++
			}
		}

		if taskEvent.StartedAt != nil && taskEvent.CompletedAt != nil {
			duration := taskEvent.CompletedAt.Sub(*taskEvent.StartedAt)
			p.TotalTaskDurationSecondsCompleted += int64(duration.Seconds())
		}
	}); err != nil {
		log.Printf("NATS ERROR updating project metrics for completed task: %v", err)
		success = false
		return
	}

	log.Println("Completed handleTaskCompleted")
}

func (s *Subscriber) handleProjectCreated(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleProjectCreated")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		p.ManagerID = projectEvent.ManagerID
	}); err != nil {
		log.Printf("NATS ERROR creating project metrics: %v", err)
		success = false
		return
	}

	log.Println("Completed handleProjectCreated")
}

func (s *Subscriber) handleProjectUpdated(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleProjectUpdated")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if projectEvent.ManagerID != "" {
			p.ManagerID = projectEvent.ManagerID
		}
	}); err != nil {
		log.Printf("NATS ERROR updating project metrics: %v", err)
		success = false
		return
	}

	log.Println("Completed handleProjectUpdated")
}

func (s *Subscriber) handleProjectMemberAdded(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleProjectMemberAdded")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		p.TeamSize++
	}); err != nil {
		log.Printf("NATS ERROR updating project metrics for member add: %v", err)
		success = false
		return
	}

	log.Println("Completed handleProjectMemberAdded")
}

func (s *Subscriber) handleProjectMemberRemoved(ctx context.Context, msg *nats.Msg) {
	log.Println("Started handleProjectMemberRemoved")
	success := true
	defer acknowledge(msg, success)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	if err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if p.TeamSize > 0 {
			p.TeamSize--
		}
	}); err != nil {
		log.Printf("NATS ERROR updating project metrics for member remove: %v", err)
		success = false
		return
	}

	log.Println("Completed handleProjectMemberRemoved")
}

func acknowledge(msg *nats.Msg, success bool) {
	if msg == nil {
		return
	}
	if success {
		if err := msg.Ack(); err != nil {
			log.Printf("NATS: failed to ack message %s: %v", msg.Subject, err)
		}
		return
	}

	if err := msg.Nak(); err != nil {
		log.Printf("NATS: failed to nak message %s: %v", msg.Subject, err)
	}
}
