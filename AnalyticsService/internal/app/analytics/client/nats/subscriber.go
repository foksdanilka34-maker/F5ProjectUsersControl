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
	_, err := s.conn.Subscribe(eventbus.EventTypeTaskCreated, s.handleTaskCreated)
	if err != nil {
		return err
	}
	_, err = s.conn.Subscribe(eventbus.EventTypeTaskAssigned, s.handleAssignTask)
	if err != nil {
		return err
	}

	_, err = s.conn.Subscribe(eventbus.EventTypeTaskCompleted, s.handleTaskCompleted)
	if err != nil {
		return err
	}

	_, err = s.conn.Subscribe(eventbus.EventTypeTaskDeleted, s.handleTaskDeleted)
	if err != nil {
		return err
	}

	_, err = s.conn.Subscribe(eventbus.EventTypeTaskStatusChanged, s.handleTaskStatusChanged)
	if err != nil {
		return err
	}

	_, err = s.conn.Subscribe(eventbus.EventTypeProjectCreated, s.handleProjectCreated)
	if err != nil {
		return err
	}
	_, err = s.conn.Subscribe(eventbus.EventTypeProjectUpdated, s.handleProjectUpdated)
	if err != nil {
		return err
	}
	_, err = s.conn.Subscribe(eventbus.EventTypeProjectMemberAdd, s.handleProjectMemberAdded)
	if err != nil {
		return err
	}
	_, err = s.conn.Subscribe(eventbus.EventTypeProjectMemberDel, s.handleProjectMemberRemoved)
	if err != nil {
		return err
	}

	log.Println("NATS subscriber started")
	log.Printf("Subscribed to topics: %s, %s, %s, %s, %s, %s, %s, %s, %s",
		eventbus.EventTypeTaskCreated,
		eventbus.EventTypeTaskAssigned,
		eventbus.EventTypeTaskCompleted,
		eventbus.EventTypeTaskDeleted,
		eventbus.EventTypeTaskStatusChanged,
		eventbus.EventTypeProjectCreated,
		eventbus.EventTypeProjectUpdated,
		eventbus.EventTypeProjectMemberAdd,
		eventbus.EventTypeProjectMemberDel)

	select {}
}

func (s *Subscriber) handleTaskCreated(msg *nats.Msg) {
	log.Println("Started handleTaskCreated")
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

	err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
		m.AssignedTasks++
		m.InProgressTasks++
	})
	if err != nil {
		log.Printf("NATS ERROR updating employee metrics for new task: %v", err)
	}

	err = s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		p.TotalTasks++
	})
	if err != nil {
		log.Printf("NATS ERROR updating project metrics for new task: %v", err)
	}
	log.Println("Completed handleTaskCreated")
}

func (s *Subscriber) handleTaskDeleted(msg *nats.Msg) {
	log.Println("Started handleTaskDeleted")
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

	err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
		if m.AssignedTasks > 0 {
			m.AssignedTasks--
		}
		if taskEvent.OldStatus == "IN_PROGRESS" {
			m.InProgressTasks--
		}
	})
	if err != nil {
		log.Printf("NATS ERROR updating employee metrics for deleted task: %v", err)
	}

	err = s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if p.TotalTasks > 0 {
			p.TotalTasks--
		}
	})
	if err != nil {
		log.Printf("NATS ERROR updating project metrics for deleted task: %v", err)
	}
	log.Println("Completed handleTaskDeleted")
}

func (s *Subscriber) handleTaskStatusChanged(msg *nats.Msg) {
	log.Println("Started handleTaskStatusChanged")
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

	err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
		switch taskEvent.OldStatus {
		case "IN_PROGRESS":
			m.InProgressTasks--
		case "DONE":
			m.CompletedTasks--
		}

		switch taskEvent.Status {
		case "IN_PROGRESS":
			m.InProgressTasks++
		case "DONE":
			m.CompletedTasks++
			if m.InProgressTasks > 0 {
				m.InProgressTasks--
			}
		}
	})
	if err != nil {
		log.Printf("NATS ERROR updating employee metrics for status change: %v", err)
	}

	err = s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		switch taskEvent.OldStatus {
		case "IN_PROGRESS":
			p.InProgressTasks--
		case "DONE":
			p.CompletedTasks--
			// This is a simplification. A more robust solution would check if the task was on time.
			if p.OnTimeCompletedTasks > 0 {
				p.OnTimeCompletedTasks--
			}
		}

		switch taskEvent.Status {
		case "IN_PROGRESS":
			p.InProgressTasks++
		case "DONE":
			// This case is handled by handleTaskCompleted
			return
		}
	})
	if err != nil {
		log.Printf("NATS ERROR updating project metrics for status change: %v", err)
	}

	log.Println("Completed handleTaskStatusChanged")
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

	if taskEvent.PrevAssigneeID != nil {
		err := s.core.UpdateEmployeeMetrics(ctx, *taskEvent.PrevAssigneeID, func(m *analytics.EmployeeMetrics) {
			m.AssignedTasks--
			if taskEvent.Status == "IN_PROGRESS" {
				m.InProgressTasks--
			}
		})
		if err != nil {
			log.Printf("NATS ERROR updating metrics for previous assignee: %v", err)
		}
	}

	if taskEvent.AssigneeID != nil {
		err = s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
			m.AssignedTasks++
			if taskEvent.Status == "IN_PROGRESS" {
				m.InProgressTasks++
			}
		})
		if err != nil {
			log.Printf("NATS ERROR updating employee metrics for new assignee: %v", err)
			return
		}
	}
	log.Println("Completed handleAssignTask")
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
	if taskEvent.AssigneeID == nil {
		log.Printf("AssigneeID is nil, skipping")
		return
	}
	err = s.core.UpdateEmployeeMetrics(ctx, *taskEvent.AssigneeID, func(m *analytics.EmployeeMetrics) {
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
	})
	if err != nil {
		log.Printf("error handling tasks complete %v", err)
	}

	err = s.core.UpdateProjectMetrics(ctx, taskEvent.ProjectID, func(p *analytics.ProjectMetrics) {
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
	})
	if err != nil {
		log.Printf("NATS ERROR updating project metrics for completed task: %v", err)
	}

	log.Println("Completed handleTaskCompleted")
}

func (s *Subscriber) handleProjectCreated(msg *nats.Msg) {
	log.Println("Started handleProjectCreated")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		p.ManagerID = projectEvent.ManagerID
	})
	if err != nil {
		log.Printf("NATS ERROR creating project metrics: %v", err)
	}
	log.Println("Completed handleProjectCreated")
}

func (s *Subscriber) handleProjectUpdated(msg *nats.Msg) {
	log.Println("Started handleProjectUpdated")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if projectEvent.ManagerID != "" {
			p.ManagerID = projectEvent.ManagerID
		}
	})
	if err != nil {
		log.Printf("NATS ERROR updating project metrics: %v", err)
	}
	log.Println("Completed handleProjectUpdated")
}

func (s *Subscriber) handleProjectMemberAdded(msg *nats.Msg) {
	log.Println("Started handleProjectMemberAdded")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		p.TeamSize++
	})
	if err != nil {
		log.Printf("NATS ERROR updating project metrics for member add: %v", err)
	}
	log.Println("Completed handleProjectMemberAdded")
}

func (s *Subscriber) handleProjectMemberRemoved(msg *nats.Msg) {
	log.Println("Started handleProjectMemberRemoved")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projectEvent := &eventbus.ProjectEvent{}
	if err := json.Unmarshal(msg.Data, projectEvent); err != nil {
		log.Printf("error unmarshaling data: %v", err)
		return
	}

	err := s.core.UpdateProjectMetrics(ctx, projectEvent.ProjectID, func(p *analytics.ProjectMetrics) {
		if p.TeamSize > 0 {
			p.TeamSize--
		}
	})
	if err != nil {
		log.Printf("NATS ERROR updating project metrics for member remove: %v", err)
	}
	log.Println("Completed handleProjectMemberRemoved")
}
