package consumer

import (
	"testing"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
)

func TestEventCategory(t *testing.T) {
	cases := map[string]string{
		"task.event.created":          dto.ExtEventTaskCreated,
		"task.event.moved":            dto.ExtEventTaskStatusChanged,
		"task.event.review_requested": dto.ExtEventTaskStatusChanged,
		"task.event.reviewed":         dto.ExtEventTaskStatusChanged,
		"task.event.approved":         dto.ExtEventTaskStatusChanged,
		"task.event.comment_added":    dto.ExtEventCommentAdded,
		"task.event.updated":          "",
		"task.event.deleted":          "",
		"task.event.assigned":         "",
	}

	for routingKey, want := range cases {
		if got := eventCategory(routingKey); got != want {
			t.Fatalf("eventCategory(%q) = %q, want %q", routingKey, got, want)
		}
	}
}
