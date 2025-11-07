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

	EventTypeTaskCreated      	= "task.created"
	EventTypeTaskUpdated       	= "task.updated"
	EventTypeTaskStatusChanged	= "task.status_changed"
	EventTypeTaskDeleted      	= "task.deleted"
	EventTypeTaskAssigned  		= "task.assigned"

	EventTypeProjectCreated     = "project.created"
	EventTypeProjectUpdated     = "project.updated"
	EventTypeProjectDeleted   	= "project.deleted"

	EventTypeProjectMemberAdd   = "project.add.member"
	EventTypeProjectMemberDel   = "project.delete.member"
)
