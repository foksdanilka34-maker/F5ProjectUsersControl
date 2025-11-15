package handlers

import (
	"log"
	"net/http"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/models"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	projectpb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProjectHandler struct {
	projectService *service.ProjectServiceClient
}

func NewProjectHandler(projectService *service.ProjectServiceClient) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProject creates a new project and assigns the current user as manager.
// @Summary      Create project
// @Tags         Projects
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        project body models.CreateProjectRequest true "Project payload"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	role := middleware.GetRoleFromContext(c)
	log.Printf("User %s (role: %s) creating new project", userID, role)

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateProject validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	protoReq := &projectpb.CreateProjectRequest{
		Name:      req.Name,
		ManagerId: userID,
	}

	if req.Description != nil {
		protoReq.Description = req.Description
	}

	if req.DueDate != nil {
		protoReq.DueDate = timestamppb.New(*req.DueDate)
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("CreateProject service error: %v", err)
		response.InternalServerError(c, "Failed to create project: "+err.Error())
		return
	}

	log.Printf("Project created successfully: id=%s, name=%s", project.Id, project.Name)
	response.Created(c, project, "Project created successfully")
}

// GetProject fetches project metadata by ID.
// @Summary      Get project
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	log.Printf("Getting project: id=%s", projectID)

	project, err := h.projectService.GetProject(c.Request.Context(), projectID)
	if err != nil {
		log.Printf("GetProject service error: %v", err)
		response.NotFound(c, "Project not found")
		return
	}

	response.Success(c, http.StatusOK, project, "Project retrieved successfully")
}

// ListProjects returns paginated projects with optional manager/status filters.
// @Summary      List projects
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        page_size   query int false "Page size"
// @Param        page_number query int false "Page number"
// @Param        manager_id  query string false "Manager filter"
// @Param        status      query int false "Project status (0-3)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	var req models.ListProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("ListProjects validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.PageNumber == 0 {
		req.PageNumber = 1
	}

	log.Printf("Listing projects: page=%d, size=%d", req.PageNumber, req.PageSize)

	protoReq := &projectpb.ListProjectsRequest{
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
	}

	if req.ManagerID != nil {
		protoReq.ManagerId = req.ManagerID
	}

	if req.Status != nil {
		status := projectpb.ProjectStatus(*req.Status)
		protoReq.Status = &status
	}

	projects, err := h.projectService.ListProjects(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("ListProjects service error: %v", err)
		response.InternalServerError(c, "Failed to list projects: "+err.Error())
		return
	}

	meta := response.PaginationMeta{PageSize: req.PageSize, PageNumber: req.PageNumber}
	response.Paginated(c, http.StatusOK, projects, meta, "Projects retrieved successfully")
}

// UpdateProject modifies project fields (name, description, status, due date).
// @Summary      Update project
// @Tags         Projects
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id      path string true "Project ID"
// @Param        project body models.UpdateProjectRequest true "Project updates"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id} [patch]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	log.Printf("User %s updating project: id=%s", userID, projectID)

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateProject validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	protoReq := &projectpb.UpdateProjectRequest{
		ProjectId: projectID,
	}

	if req.Name != nil {
		protoReq.Name = req.Name
	}
	if req.Description != nil {
		protoReq.Description = req.Description
	}
	if req.Status != nil {
		status := projectpb.ProjectStatus(*req.Status)
		protoReq.Status = &status
	}
	if req.DueDate != nil {
		protoReq.DueDate = timestamppb.New(*req.DueDate)
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("UpdateProject service error: %v", err)
		response.InternalServerError(c, "Failed to update project: "+err.Error())
		return
	}

	log.Printf("Project updated successfully: id=%s", project.Id)
	response.Success(c, http.StatusOK, project, "Project updated successfully")
}

// DeleteProject removes a project by ID.
// @Summary      Delete project
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	log.Printf("User %s deleting project: id=%s", userID, projectID)

	err := h.projectService.DeleteProject(c.Request.Context(), projectID)
	if err != nil {
		log.Printf("DeleteProject service error: %v", err)
		response.InternalServerError(c, "Failed to delete project: "+err.Error())
		return
	}

	log.Printf("Project deleted successfully: id=%s", projectID)
	response.Success(c, http.StatusOK, nil, "Project deleted successfully")
}

// CreateTask adds a task to a project.
// @Summary      Create task
// @Tags         Projects
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Project ID"
// @Param        task body models.CreateTaskRequest true "Task payload"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/tasks [post]
func (h *ProjectHandler) CreateTask(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")

	log.Printf("User %s creating task in project: %s", userID, projectID)

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateTask validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	protoReq := &projectpb.CreateTaskRequest{
		ProjectId:   projectID,
		Title:       req.Title,
		Description: req.Description,
		CreatorId:   userID,
		DueDate:     timestamppb.New(req.DueDate),
	}

	if req.Priority != nil {
		priority := projectpb.TaskPriority(*req.Priority)
		protoReq.Priority = &priority
	}

	if req.AssigneeID != nil {
		protoReq.AssigneeId = req.AssigneeID
	}

	task, err := h.projectService.CreateTask(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("CreateTask service error: %v", err)
		response.InternalServerError(c, "Failed to create task: "+err.Error())
		return
	}

	log.Printf("Task created successfully: id=%s, title=%s", task.Id, task.Title)
	response.Created(c, task, "Task created successfully")
}

// GetTask fetches a specific task within a project.
// @Summary      Get task
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id     path string true "Project ID"
// @Param        taskId path string true "Task ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/projects/{id}/tasks/{taskId} [get]
func (h *ProjectHandler) GetTask(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "Task ID is required")
		return
	}

	log.Printf("Getting task: projectID=%s, taskID=%s", projectID, taskID)

	task, err := h.projectService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("GetTask service error: %v", err)
		response.NotFound(c, "Task not found")
		return
	}

	if task.ProjectId != projectID {
		response.BadRequest(c, "Task does not belong to the specified project")
		return
	}

	response.Success(c, http.StatusOK, task, "Task retrieved successfully")
}

// UpdateTask modifies fields of an existing task.
// @Summary      Update task
// @Tags         Projects
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id     path string true "Project ID"
// @Param        taskId path string true "Task ID"
// @Param        task   body models.UpdateTaskRequest true "Task updates"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/tasks/{taskId} [patch]
func (h *ProjectHandler) UpdateTask(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "Task ID is required")
		return
	}

	log.Printf("User %s updating task: projectID=%s, taskID=%s", userID, projectID, taskID)

	existingTask, err := h.projectService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("UpdateTask: failed to get existing task: %v", err)
		response.NotFound(c, "Task not found")
		return
	}
	if existingTask.ProjectId != projectID {
		response.BadRequest(c, "Task does not belong to the specified project")
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateTask validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	protoReq := &projectpb.UpdateTaskRequest{
		TaskId: taskID,
	}

	if req.Title != nil {
		protoReq.Title = req.Title
	}
	if req.Description != nil {
		protoReq.Description = req.Description
	}
	if req.Status != nil {
		status := projectpb.TaskStatus(*req.Status)
		protoReq.Status = &status
	}
	if req.Priority != nil {
		priority := projectpb.TaskPriority(*req.Priority)
		protoReq.Priority = &priority
	}
	if req.AssigneeID != nil {
		protoReq.AssigneeId = req.AssigneeID
	}
	if req.DueDate != nil {
		protoReq.DueDate = timestamppb.New(*req.DueDate)
	}

	task, err := h.projectService.UpdateTask(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("UpdateTask service error: %v", err)
		response.InternalServerError(c, "Failed to update task: "+err.Error())
		return
	}

	log.Printf("Task updated successfully: id=%s", task.Id)
	response.Success(c, http.StatusOK, task, "Task updated successfully")
}

// DeleteTask removes a task from a project.
// @Summary      Delete task
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id     path string true "Project ID"
// @Param        taskId path string true "Task ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/tasks/{taskId} [delete]
func (h *ProjectHandler) DeleteTask(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "Task ID is required")
		return
	}

	log.Printf("User %s deleting task: projectID=%s, taskID=%s", userID, projectID, taskID)

	existingTask, err := h.projectService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("DeleteTask: failed to get existing task: %v", err)
		response.NotFound(c, "Task not found")
		return
	}
	if existingTask.ProjectId != projectID {
		response.BadRequest(c, "Task does not belong to the specified project")
		return
	}

	err = h.projectService.DeleteTask(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("DeleteTask service error: %v", err)
		response.InternalServerError(c, "Failed to delete task: "+err.Error())
		return
	}

	log.Printf("Task deleted successfully: id=%s", taskID)
	response.Success(c, http.StatusOK, nil, "Task deleted successfully")
}

// MoveTask changes a task status/order within a project.
// @Summary      Move task
// @Tags         Projects
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id     path string true "Project ID"
// @Param        taskId path string true "Task ID"
// @Param        move   body models.MoveTaskRequest true "Move payload"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/tasks/{taskId}/move [post]
func (h *ProjectHandler) MoveTask(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "Task ID is required")
		return
	}

	log.Printf("User %s moving task: projectID=%s, taskID=%s", userID, projectID, taskID)

	existingTask, err := h.projectService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("MoveTask: failed to get existing task: %v", err)
		response.NotFound(c, "Task not found")
		return
	}
	if existingTask.ProjectId != projectID {
		response.BadRequest(c, "Task does not belong to the specified project")
		return
	}

	var req models.MoveTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("MoveTask validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	protoReq := &projectpb.MoveTaskRequest{
		TaskId:        taskID,
		NewStatus:     projectpb.TaskStatus(req.NewStatus),
		NewOrderIndex: req.NewOrderIndex,
	}

	task, err := h.projectService.MoveTask(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("MoveTask service error: %v", err)
		response.InternalServerError(c, "Failed to move task: "+err.Error())
		return
	}

	log.Printf("Task moved successfully: id=%s", task.Id)
	response.Success(c, http.StatusOK, task, "Task moved successfully")
}

// AssignTask delegates a task to a team member.
// @Summary      Assign task
// @Tags         Projects
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id        path string true "Project ID"
// @Param        taskId    path string true "Task ID"
// @Param        assignee  body models.AssignTaskRequest true "Assignee payload"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/tasks/{taskId}/assign [post]
func (h *ProjectHandler) AssignTask(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "Task ID is required")
		return
	}

	log.Printf("User %s assigning task: projectID=%s, taskID=%s", userID, projectID, taskID)

	existingTask, err := h.projectService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("AssignTask: failed to get existing task: %v", err)
		response.NotFound(c, "Task not found")
		return
	}
	if existingTask.ProjectId != projectID {
		response.BadRequest(c, "Task does not belong to the specified project")
		return
	}

	var req models.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("AssignTask validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	task, err := h.projectService.AssignTask(c.Request.Context(), taskID, req.AssigneeID)
	if err != nil {
		log.Printf("AssignTask service error: %v", err)
		response.InternalServerError(c, "Failed to assign task: "+err.Error())
		return
	}

	log.Printf("Task assigned successfully: id=%s, assigneeID=%s", task.Id, req.AssigneeID)
	response.Success(c, http.StatusOK, task, "Task assigned successfully")
}

// ListTasksByProject lists tasks for a project with optional filters.
// @Summary      List project tasks
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id          path   string true  "Project ID"
// @Param        status      query  int32  false "Task status filter"
// @Param        assignee_id query  string false "Assignee filter"
// @Param        priority    query  int32  false "Priority filter"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/tasks [get]
func (h *ProjectHandler) ListTasksByProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	var req models.ListTasksByProjectRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("ListTasksByProject validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	log.Printf("Listing tasks for project: id=%s", projectID)

	protoReq := &projectpb.ListTasksByProjectRequest{
		ProjectId: projectID,
	}

	if req.Status != nil {
		status := projectpb.TaskStatus(*req.Status)
		protoReq.Status = &status
	}
	if req.AssigneeID != nil {
		protoReq.AssigneeId = req.AssigneeID
	}
	if req.Priority != nil {
		priority := projectpb.TaskPriority(*req.Priority)
		protoReq.Priority = &priority
	}

	tasks, err := h.projectService.ListTasksByProject(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("ListTasksByProject service error: %v", err)
		response.InternalServerError(c, "Failed to list tasks: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, tasks, "Tasks retrieved successfully")
}

// AddMemberToProject links an employee to a project team.
// @Summary      Add project member
// @Tags         Projects
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id     path string true "Project ID"
// @Param        member body models.AddMemberToProjectRequest true "Member payload"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/members [post]
func (h *ProjectHandler) AddMemberToProject(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	log.Printf("User %s adding member to project: id=%s", userID, projectID)

	var req models.AddMemberToProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("AddMemberToProject validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	err := h.projectService.AddMemberToProject(c.Request.Context(), projectID, req.UserID)
	if err != nil {
		log.Printf("AddMemberToProject service error: %v", err)
		response.InternalServerError(c, "Failed to add member to project: "+err.Error())
		return
	}

	log.Printf("Member added to project successfully: projectID=%s, userID=%s", projectID, req.UserID)
	response.Success(c, http.StatusOK, nil, "Member added to project successfully")
}

// RemoveMemberFromProject removes a project member.
// @Summary      Remove project member
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id       path string true "Project ID"
// @Param        memberId path string true "Member ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/members/{memberId} [delete]
func (h *ProjectHandler) RemoveMemberFromProject(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	projectID := c.Param("id")
	memberID := c.Param("memberId")

	if projectID == "" || memberID == "" {
		response.BadRequest(c, "Project ID and Member ID are required")
		return
	}

	log.Printf("User %s removing member from project: projectID=%s, memberID=%s", userID, projectID, memberID)

	err := h.projectService.RemoveMemberFromProject(c.Request.Context(), projectID, memberID)
	if err != nil {
		log.Printf("RemoveMemberFromProject service error: %v", err)
		response.InternalServerError(c, "Failed to remove member from project: "+err.Error())
		return
	}

	log.Printf("Member removed from project successfully: projectID=%s, memberID=%s", projectID, memberID)
	response.Success(c, http.StatusOK, nil, "Member removed from project successfully")
}

// ListProjectMembers retrieves members assigned to a project.
// @Summary      List project members
// @Tags         Projects
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/projects/{id}/members [get]
func (h *ProjectHandler) ListProjectMembers(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	log.Printf("Listing members for project: id=%s", projectID)

	members, err := h.projectService.ListProjectMembers(c.Request.Context(), projectID)
	if err != nil {
		log.Printf("ListProjectMembers service error: %v", err)
		response.InternalServerError(c, "Failed to list project members: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, members, "Project members retrieved successfully")
}
