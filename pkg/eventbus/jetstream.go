package eventbus

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

var jetStreamStreamConfigs = []nats.StreamConfig{
	{
		Name: "employee-events",
		Subjects: []string{
			EmployeeCreatedEventTopic,
			EmployeeUpdatedEventTopic,
			EmployeeDeletedEventTopic,
		},
		Storage:  nats.FileStorage,
		Replicas: 1,
	},
	{
		Name: "project-events",
		Subjects: []string{
			EventTypeTaskCreated,
			EventTypeTaskStatusChanged,
			EventTypeTaskDeleted,
			EventTypeTaskAssigned,
			EventTypeTaskCompleted,
			EventTypeProjectCreated,
			EventTypeProjectUpdated,
			EventTypeProjectDeleted,
			EventTypeProjectMemberAdd,
			EventTypeProjectMemberDel,
			ProjectTasksTopic,
			ProjectProjectsTopic,
		},
		Storage:  nats.FileStorage,
		Replicas: 1,
	},
	{
		Name: "login-commands",
		Subjects: []string{
			LoginDeactivateUserCommandTopic,
		},
		Storage:  nats.FileStorage,
		Replicas: 1,
	},
}

func EnsureJetStreamStreams(js nats.JetStreamContext) error {
	for _, cfg := range jetStreamStreamConfigs {
		if _, err := js.StreamInfo(cfg.Name); err == nil {
			continue
		} else if err != nil && err != nats.ErrStreamNotFound {
			return fmt.Errorf("failed to query stream %s: %w", cfg.Name, err)
		}

		if _, err := js.AddStream(&cfg); err != nil && err != nats.ErrStreamNameAlreadyInUse {
			return fmt.Errorf("failed to create stream %s: %w", cfg.Name, err)
		}
	}

	return nil
}
