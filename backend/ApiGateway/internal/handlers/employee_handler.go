package handlers

import (
	"log"
	"net/http"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/models"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	employeepb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/employee_service"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EmployeeHandler struct {
	employeeService *service.EmployeeServiceClient
}

func NewEmployeeHandler(employeeService *service.EmployeeServiceClient) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
	}
}

// CreateProfile creates a new employee profile and returns the created record.
// @Summary      Create employee profile
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        profile body models.CreateProfileRequest true "Profile payload"
// @Success      201  {object} response.Response
// @Failure      400  {object} response.Response
// @Failure      500  {object} response.Response
// @Router       /api/v1/employees/profiles [post]
func (h *EmployeeHandler) CreateProfile(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	log.Printf("Admin %s creating new profile", adminID)

	var req models.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateProfile validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	protoReq := &employeepb.CreateProfileRequest{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		PositionId: req.PositionID,
		Email:      req.Email,
		HireDate:   timestamppb.New(req.HireDate),
		Login:      req.Login,
		Password:   req.Password,
		Role:       req.Role,
	}

	if req.DepartmentID != nil {
		protoReq.DepartmentId = req.DepartmentID
	}

	profile, err := h.employeeService.CreateProfile(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("CreateProfile service error: %v", err)
		response.InternalServerError(c, "Failed to create profile: "+err.Error())
		return
	}

	log.Printf("Profile created successfully: id=%s", profile.Id)
	response.Created(c, profile, "Profile created successfully")
}

// GetProfile returns a profile by its ID.
// @Summary      Get employee profile
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Profile ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/employees/profiles/{id} [get]
func (h *EmployeeHandler) GetProfile(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "User ID is required")
		return
	}

	log.Printf("Getting profile: id=%s", userID)

	profile, err := h.employeeService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		log.Printf("GetProfile service error: %v", err)
		response.NotFound(c, "Profile not found")
		return
	}

	response.Success(c, http.StatusOK, profile, "Profile retrieved successfully")
}

// ListProfiles returns paginated employee profiles with optional filtering.
// @Summary      List employee profiles
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Param        page_size   query int false "Page size"
// @Param        page_number query int false "Page number"
// @Param        department_id query string false "Department filter"
// @Param        position_id query string false "Position filter"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/profiles [get]
func (h *EmployeeHandler) ListProfiles(c *gin.Context) {
	var req models.ListProfilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("ListProfiles validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.PageNumber == 0 {
		req.PageNumber = 1
	}

	log.Printf("Listing profiles: page=%d, size=%d", req.PageNumber, req.PageSize)

	profiles, err := h.employeeService.ListProfiles(
		c.Request.Context(),
		req.PageSize,
		req.PageNumber,
		req.DepartmentID,
		req.PositionID,
	)
	if err != nil {
		log.Printf("ListProfiles service error: %v", err)
		response.InternalServerError(c, "Failed to list profiles: "+err.Error())
		return
	}

	meta := response.PaginationMeta{PageSize: req.PageSize, PageNumber: req.PageNumber}
	response.Paginated(c, http.StatusOK, profiles, meta, "Profiles retrieved successfully")
}

// UpdateProfile applies partial updates to an existing profile.
// @Summary      Update employee profile
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id      path string true "Profile ID"
// @Param        profile body models.UpdateProfileRequest true "Profile updates"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/profiles/{id} [patch]
func (h *EmployeeHandler) UpdateProfile(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "User ID is required")
		return
	}

	log.Printf("Admin %s updating profile: id=%s", adminID, userID)

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateProfile validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	protoReq := &employeepb.UpdateProfileRequest{
		UserId: userID,
	}

	if req.FirstName != nil {
		protoReq.FirstName = req.FirstName
	}
	if req.LastName != nil {
		protoReq.LastName = req.LastName
	}
	if req.PositionID != nil {
		protoReq.PositionId = req.PositionID
	}
	if req.Email != nil {
		protoReq.Email = req.Email
	}
	if req.DepartmentID != nil {
		protoReq.DepartmentId = req.DepartmentID
	}
	if req.AvatarURL != nil {
		protoReq.AvatarUrl = req.AvatarURL
	}

	profile, err := h.employeeService.UpdateProfile(c.Request.Context(), protoReq)
	if err != nil {
		log.Printf("UpdateProfile service error: %v", err)
		response.InternalServerError(c, "Failed to update profile: "+err.Error())
		return
	}

	log.Printf("Profile updated successfully: id=%s", profile.Id)
	response.Success(c, http.StatusOK, profile, "Profile updated successfully")
}

// ChangeUserStatus toggles the status of an employee profile.
// @Summary      Change user status
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Profile ID"
// @Param        status body models.ChangeUserStatusRequest true "Status payload"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/profiles/{id}/status [patch]
func (h *EmployeeHandler) ChangeUserStatus(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "User ID is required")
		return
	}

	log.Printf("Admin %s changing user status: id=%s", adminID, userID)

	var req models.ChangeUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ChangeUserStatus validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	err := h.employeeService.ChangeUserStatusProfile(c.Request.Context(), userID, *req.Status)
	if err != nil {
		log.Printf("ChangeUserStatus service error: %v", err)
		response.InternalServerError(c, "Failed to change user status: "+err.Error())
		return
	}

	statusStr := "activated"
	if !*req.Status {
		statusStr = "deactivated"
	}

	log.Printf("User status changed successfully: id=%s, status=%s", userID, statusStr)
	response.Success(c, http.StatusOK, nil, "User status changed successfully")
}

// CreateDepartment adds a new department record.
// @Summary      Create department
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        department body models.CreateDepartmentRequest true "Department payload"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/departments [post]
func (h *EmployeeHandler) CreateDepartment(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	log.Printf("Admin %s creating new department", adminID)

	var req models.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateDepartment validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	department, err := h.employeeService.CreateDepartment(c.Request.Context(), req.Name)
	if err != nil {
		log.Printf("CreateDepartment service error: %v", err)
		response.InternalServerError(c, "Failed to create department: "+err.Error())
		return
	}

	log.Printf("Department created successfully: id=%s, name=%s", department.Id, department.Name)
	response.Created(c, department, "Department created successfully")
}

// GetDepartment retrieves a department by ID.
// @Summary      Get department
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Department ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/employees/departments/{id} [get]
func (h *EmployeeHandler) GetDepartment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Department ID is required")
		return
	}

	log.Printf("Getting department: id=%s", id)

	department, err := h.employeeService.GetDepartment(c.Request.Context(), id)
	if err != nil {
		log.Printf("GetDepartment service error: %v", err)
		response.NotFound(c, "Department not found")
		return
	}

	response.Success(c, http.StatusOK, department, "Department retrieved successfully")
}

// ListDepartments returns all departments.
// @Summary      List departments
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/departments [get]
func (h *EmployeeHandler) ListDepartments(c *gin.Context) {
	log.Printf("Listing all departments")

	departments, err := h.employeeService.ListDepartments(c.Request.Context())
	if err != nil {
		log.Printf("ListDepartments service error: %v", err)
		response.InternalServerError(c, "Failed to list departments: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, departments, "Departments retrieved successfully")
}

// UpdateDepartment renames an existing department.
// @Summary      Update department
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Department ID"
// @Param        department body models.UpdateDepartmentRequest true "Department payload"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/departments/{id} [put]
func (h *EmployeeHandler) UpdateDepartment(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Department ID is required")
		return
	}

	log.Printf("Admin %s updating department: id=%s", adminID, id)

	var req models.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateDepartment validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	department, err := h.employeeService.UpdateDepartment(c.Request.Context(), id, req.Name)
	if err != nil {
		log.Printf("UpdateDepartment service error: %v", err)
		response.InternalServerError(c, "Failed to update department: "+err.Error())
		return
	}

	log.Printf("Department updated successfully: id=%s, name=%s", department.Id, department.Name)
	response.Success(c, http.StatusOK, department, "Department updated successfully")
}

// DeleteDepartment removes a department by ID.
// @Summary      Delete department
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Department ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/departments/{id} [delete]
func (h *EmployeeHandler) DeleteDepartment(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Department ID is required")
		return
	}

	log.Printf("Admin %s deleting department: id=%s", adminID, id)

	err := h.employeeService.DeleteDepartment(c.Request.Context(), id)
	if err != nil {
		log.Printf("DeleteDepartment service error: %v", err)
		response.InternalServerError(c, "Failed to delete department: "+err.Error())
		return
	}

	log.Printf("Department deleted successfully: id=%s", id)
	response.Success(c, http.StatusOK, nil, "Department deleted successfully")
}

// CreatePosition adds a new position.
// @Summary      Create position
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        position body models.CreatePositionRequest true "Position payload"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/positions [post]
func (h *EmployeeHandler) CreatePosition(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	log.Printf("Admin %s creating new position", adminID)

	var req models.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreatePosition validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	position, err := h.employeeService.CreatePosition(c.Request.Context(), req.Name)
	if err != nil {
		log.Printf("CreatePosition service error: %v", err)
		response.InternalServerError(c, "Failed to create position: "+err.Error())
		return
	}

	log.Printf("Position created successfully: id=%s, name=%s", position.Id, position.Name)
	response.Created(c, position, "Position created successfully")
}

// GetPosition retrieves a position by ID.
// @Summary      Get position
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Position ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/employees/positions/{id} [get]
func (h *EmployeeHandler) GetPosition(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Position ID is required")
		return
	}

	log.Printf("Getting position: id=%s", id)

	position, err := h.employeeService.GetPosition(c.Request.Context(), id)
	if err != nil {
		log.Printf("GetPosition service error: %v", err)
		response.NotFound(c, "Position not found")
		return
	}

	response.Success(c, http.StatusOK, position, "Position retrieved successfully")
}

// ListPositions returns all available positions.
// @Summary      List positions
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/positions [get]
func (h *EmployeeHandler) ListPositions(c *gin.Context) {
	log.Printf("Listing all positions")

	positions, err := h.employeeService.ListPositions(c.Request.Context())
	if err != nil {
		log.Printf("ListPositions service error: %v", err)
		response.InternalServerError(c, "Failed to list positions: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, positions, "Positions retrieved successfully")
}

// UpdatePosition renames an existing position.
// @Summary      Update position
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Position ID"
// @Param        position body models.UpdatePositionRequest true "Position payload"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/positions/{id} [put]
func (h *EmployeeHandler) UpdatePosition(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Position ID is required")
		return
	}

	log.Printf("Admin %s updating position: id=%s", adminID, id)

	var req models.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdatePosition validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	position, err := h.employeeService.UpdatePosition(c.Request.Context(), id, req.Name)
	if err != nil {
		log.Printf("UpdatePosition service error: %v", err)
		response.InternalServerError(c, "Failed to update position: "+err.Error())
		return
	}

	log.Printf("Position updated successfully: id=%s, name=%s", position.Id, position.Name)
	response.Success(c, http.StatusOK, position, "Position updated successfully")
}

// DeletePosition removes a position by ID.
// @Summary      Delete position
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Position ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/positions/{id} [delete]
func (h *EmployeeHandler) DeletePosition(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Position ID is required")
		return
	}

	log.Printf("Admin %s deleting position: id=%s", adminID, id)

	err := h.employeeService.DeletePosition(c.Request.Context(), id)
	if err != nil {
		log.Printf("DeletePosition service error: %v", err)
		response.InternalServerError(c, "Failed to delete position: "+err.Error())
		return
	}

	log.Printf("Position deleted successfully: id=%s", id)
	response.Success(c, http.StatusOK, nil, "Position deleted successfully")
}

// CreateSkill registers a new skill dimension.
// @Summary      Create skill
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        skill body models.CreateSkillRequest true "Skill payload"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/skills [post]
func (h *EmployeeHandler) CreateSkill(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	log.Printf("Admin %s creating new skill", adminID)

	var req models.CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateSkill validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	skill, err := h.employeeService.CreateSkill(c.Request.Context(), req.Name)
	if err != nil {
		log.Printf("CreateSkill service error: %v", err)
		response.InternalServerError(c, "Failed to create skill: "+err.Error())
		return
	}

	log.Printf("Skill created successfully: id=%s, name=%s", skill.Id, skill.Name)
	response.Created(c, skill, "Skill created successfully")
}

// ListSkills returns all available skills.
// @Summary      List skills
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/skills [get]
func (h *EmployeeHandler) ListSkills(c *gin.Context) {
	log.Printf("Listing all skills")

	skills, err := h.employeeService.ListSkills(c.Request.Context())
	if err != nil {
		log.Printf("ListSkills service error: %v", err)
		response.InternalServerError(c, "Failed to list skills: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, skills, "Skills retrieved successfully")
}

// AddSkillToEmployee attaches a skill to an employee profile.
// @Summary      Add skill to employee
// @Tags         Employees
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Profile ID"
// @Param        skill body models.AddSkillToEmployeeRequest true "Skill payload"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/profiles/{id}/skills [post]
func (h *EmployeeHandler) AddSkillToEmployee(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "Employee ID is required")
		return
	}

	log.Printf("Admin %s adding skill to employee: id=%s", adminID, employeeID)

	var req models.AddSkillToEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("AddSkillToEmployee validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	err := h.employeeService.AddSkillToEmployee(c.Request.Context(), employeeID, req.SkillID)
	if err != nil {
		log.Printf("AddSkillToEmployee service error: %v", err)
		response.InternalServerError(c, "Failed to add skill to employee: "+err.Error())
		return
	}

	log.Printf("Skill added to employee successfully: employeeID=%s, skillID=%s", employeeID, req.SkillID)
	response.Success(c, http.StatusOK, nil, "Skill added to employee successfully")
}

// RemoveSkillFromEmployee removes a skill from an employee profile.
// @Summary      Remove skill from employee
// @Tags         Employees
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Profile ID"
// @Param        skillId path string true "Skill ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/employees/profiles/{id}/skills/{skillId} [delete]
func (h *EmployeeHandler) RemoveSkillFromEmployee(c *gin.Context) {
	adminID := middleware.GetUserIDFromContext(c)
	employeeID := c.Param("id")
	skillID := c.Param("skillId")

	if employeeID == "" || skillID == "" {
		response.BadRequest(c, "Employee ID and Skill ID are required")
		return
	}

	log.Printf("Admin %s removing skill from employee: employeeID=%s, skillID=%s", adminID, employeeID, skillID)

	err := h.employeeService.RemoveSkillFromEmployee(c.Request.Context(), employeeID, skillID)
	if err != nil {
		log.Printf("RemoveSkillFromEmployee service error: %v", err)
		response.InternalServerError(c, "Failed to remove skill from employee: "+err.Error())
		return
	}

	log.Printf("Skill removed from employee successfully: employeeID=%s, skillID=%s", employeeID, skillID)
	response.Success(c, http.StatusOK, nil, "Skill removed from employee successfully")
}
