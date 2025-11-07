package eventbus

const (
	ProjectTasksTopic    = "project.tasks"
	ProjectProjectsTopic = "project.projects"

	EmployeeTaskAssignedTopic  = "events.employee.task.assigned"
	EmployeeTaskCompletedTopic = "events.employee.task.completed"
	ProjectTaskStatusTopic     = "events.project.task.status_changed"

	LoginDeactivateUserCommandTopic = "login.command.deactivate"

	EmployeeCreatedEventTopic = "employee.event.created"
	EmployeeUpdatedEventTopic = "employee.event.updated"
	EmployeeDeletedEventTopic = "employee.event.deleted"
)
