package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/identity"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProfileHTTPHandler struct {
	client pb.IdentityServiceClient
}

func NewProfileHTTPHandler(client pb.IdentityServiceClient) *ProfileHTTPHandler {
	return &ProfileHTTPHandler{client: client}
}

func (h *ProfileHTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	// Profiles
	mux.HandleFunc("POST /api/v1/profiles", h.Create)
	mux.HandleFunc("GET /api/v1/profiles", h.List)
	mux.HandleFunc("GET /api/v1/profiles/{id}", h.Get)
	mux.HandleFunc("PUT /api/v1/profiles/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/profiles/{id}", h.Delete)

	// Skills для профилей
	mux.HandleFunc("POST /api/v1/profiles/{id}/skills", h.AddSkillToProfile)
	mux.HandleFunc("DELETE /api/v1/profiles/{id}/skills/{skillId}", h.RemoveSkillFromProfile)

	// Departments
	mux.HandleFunc("POST /api/v1/departments", h.CreateDepartment)
	mux.HandleFunc("GET /api/v1/departments", h.ListDepartments)
	mux.HandleFunc("GET /api/v1/departments/{id}", h.GetDepartment)
	mux.HandleFunc("PUT /api/v1/departments/{id}", h.UpdateDepartment)
	mux.HandleFunc("DELETE /api/v1/departments/{id}", h.DeleteDepartment)

	// Positions
	mux.HandleFunc("POST /api/v1/positions", h.CreatePosition)
	mux.HandleFunc("GET /api/v1/positions", h.ListPositions)
	mux.HandleFunc("PUT /api/v1/positions/{id}", h.UpdatePosition)
	mux.HandleFunc("DELETE /api/v1/positions/{id}", h.DeletePosition)

	// Skills
	mux.HandleFunc("POST /api/v1/skills", h.CreateSkill)
	mux.HandleFunc("GET /api/v1/skills", h.ListSkills)
}

func (h *ProfileHTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		PositionID   int64  `json:"position_id"`
		DepartmentID *int64 `json:"department_id"`
		Email        string `json:"email"`
		Login        string `json:"login"`
		Password     string `json:"password"`
		Role         string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("profile create decode error:", err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.PositionID == 0 || req.Email == "" || req.Login == "" || req.Password == "" {
		http.Error(w, `{"error":"required fields missing"}`, http.StatusBadRequest)
		return
	}

	pbReq := &pb.CreateProfileRequest{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		PositionId: req.PositionID,
		Email:      req.Email,
		Login:      req.Login,
		Password:   req.Password,
		Role:       req.Role,
	}
	if req.DepartmentID != nil {
		pbReq.DepartmentId = req.DepartmentID
	}

	profile, err := h.client.CreateProfile(r.Context(), pbReq)
	if err != nil {
		log.Println("profile create error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, profileToMap(profile))
}

func (h *ProfileHTTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	profile, err := h.client.GetProfile(r.Context(), &pb.GetProfileRequest{UserId: id})
	if err != nil {
		log.Println("profile get error:", err)
		http.Error(w, `{"error":"profile not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, profileToMap(profile))
}

func (h *ProfileHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	pageNumber, _ := strconv.Atoi(query.Get("page_number"))
	if pageSize == 0 {
		pageSize = 20
	}
	if pageNumber == 0 {
		pageNumber = 1
	}

	var departmentID, positionID int64
	if deptStr := query.Get("department_id"); deptStr != "" {
		departmentID, _ = parseID(deptStr)
	}
	if posStr := query.Get("position_id"); posStr != "" {
		positionID, _ = parseID(posStr)
	}

	resp, err := h.client.ListProfiles(r.Context(), &pb.ListProfilesRequest{
		PageSize:     int32(pageSize),
		PageNumber:   int32(pageNumber),
		DepartmentId: departmentID,
		PositionId:   positionID,
	})
	if err != nil {
		log.Println("profile list error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	profiles := make([]map[string]interface{}, len(resp.Profiles))
	for i, p := range resp.Profiles {
		profiles[i] = profileToMap(p)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"profiles":    profiles,
		"total_count": resp.TotalCount,
	})
}

func (h *ProfileHTTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		FirstName    *string `json:"first_name"`
		LastName     *string `json:"last_name"`
		PositionID   *int64  `json:"position_id"`
		DepartmentID *int64  `json:"department_id"`
		Email        *string `json:"email"`
		AvatarURL    *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	profile, err := h.client.UpdateProfile(r.Context(), &pb.UpdateProfileRequest{
		UserId:       id,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PositionId:   req.PositionID,
		DepartmentId: req.DepartmentID,
		Email:        req.Email,
		AvatarUrl:    req.AvatarURL,
	})
	if err != nil {
		log.Println("profile update error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, profileToMap(profile))
}

func (h *ProfileHTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.ChangeUserStatus(r.Context(), &pb.ChangeUserStatusRequest{
		UserId:   id,
		IsActive: false,
	})
	if err != nil {
		log.Println("profile delete error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "profile deleted"})
}

// Department handlers
func (h *ProfileHTTPHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}

	dept, err := h.client.CreateDepartment(r.Context(), &pb.CreateDepartmentRequest{Name: req.Name})
	if err != nil {
		log.Println("department create error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": dept.Id, "name": dept.Name})
}

func (h *ProfileHTTPHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	dept, err := h.client.GetDepartment(r.Context(), &pb.GetDepartmentRequest{Id: id})
	if err != nil {
		http.Error(w, `{"error":"department not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"id": dept.Id, "name": dept.Name})
}

func (h *ProfileHTTPHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListDepartments(r.Context(), &emptypb.Empty{})
	if err != nil {
		log.Println("departments list error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	departments := make([]map[string]interface{}, len(resp.Departments))
	for i, d := range resp.Departments {
		departments[i] = map[string]interface{}{"id": d.Id, "name": d.Name}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"departments": departments})
}

func (h *ProfileHTTPHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}

	dept, err := h.client.UpdateDepartment(r.Context(), &pb.UpdateDepartmentRequest{Id: id, Name: req.Name})
	if err != nil {
		log.Println("department update error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"id": dept.Id, "name": dept.Name})
}

func (h *ProfileHTTPHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteDepartment(r.Context(), &pb.DeleteDepartmentRequest{Id: id})
	if err != nil {
		log.Println("department delete error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "department deleted"})
}

// Position handlers
func (h *ProfileHTTPHandler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}

	pos, err := h.client.CreatePosition(r.Context(), &pb.CreatePositionRequest{Name: req.Name})
	if err != nil {
		log.Println("position create error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": pos.Id, "name": pos.Name})
}

func (h *ProfileHTTPHandler) ListPositions(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListPositions(r.Context(), &emptypb.Empty{})
	if err != nil {
		log.Println("positions list error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	positions := make([]map[string]interface{}, len(resp.Positions))
	for i, p := range resp.Positions {
		positions[i] = map[string]interface{}{"id": p.Id, "name": p.Name}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"positions": positions})
}

func (h *ProfileHTTPHandler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}

	pos, err := h.client.UpdatePosition(r.Context(), &pb.UpdatePositionRequest{Id: id, Name: req.Name})
	if err != nil {
		log.Println("position update error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"id": pos.Id, "name": pos.Name})
}

func (h *ProfileHTTPHandler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.DeletePosition(r.Context(), &pb.DeletePositionRequest{Id: id})
	if err != nil {
		log.Println("position delete error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "position deleted"})
}

// Skill handlers
func (h *ProfileHTTPHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}

	skill, err := h.client.CreateSkill(r.Context(), &pb.CreateSkillRequest{Name: req.Name})
	if err != nil {
		log.Println("skill create error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": skill.Id, "name": skill.Name})
}

func (h *ProfileHTTPHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListSkills(r.Context(), &emptypb.Empty{})
	if err != nil {
		log.Println("skills list error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	skills := make([]map[string]interface{}, len(resp.Skills))
	for i, s := range resp.Skills {
		skills[i] = map[string]interface{}{"id": s.Id, "name": s.Name}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"skills": skills})
}

func (h *ProfileHTTPHandler) AddSkillToProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		SkillID int64 `json:"skill_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SkillID == 0 {
		http.Error(w, `{"error":"skill_id required"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.AddSkillToEmployee(r.Context(), &pb.AddSkillToEmployeeRequest{
		EmployeeId: id,
		SkillId:    req.SkillID,
	})
	if err != nil {
		log.Println("add skill error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "skill added"})
}

func (h *ProfileHTTPHandler) RemoveSkillFromProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	skillIDStr := r.PathValue("skillId")
	skillID, err := parseID(skillIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid skill_id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.RemoveSkillFromEmployee(r.Context(), &pb.RemoveSkillFromEmployeeRequest{
		EmployeeId: id,
		SkillId:    skillID,
	})
	if err != nil {
		log.Println("remove skill error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "skill removed"})
}

func profileToMap(p *pb.Profile) map[string]interface{} {
	resp := map[string]interface{}{
		"id":          p.Id,
		"first_name":  p.FirstName,
		"last_name":   p.LastName,
		"position_id": p.PositionId,
		"email":       p.Email,
		"avatar_url":  p.AvatarUrl,
		"login":       p.Login,
		"role":        p.Role,
		"is_active":   p.IsActive,
	}

	if p.HireDate != nil {
		resp["hire_date"] = p.HireDate.AsTime().Format("2006-01-02")
	}
	if p.CreatedAt != nil {
		resp["created_at"] = p.CreatedAt.AsTime()
	}
	if p.UpdatedAt != nil {
		resp["updated_at"] = p.UpdatedAt.AsTime()
	}
	if p.Department != nil {
		resp["department"] = map[string]interface{}{
			"id":   p.Department.Id,
			"name": p.Department.Name,
		}
	}
	if len(p.Skills) > 0 {
		skills := make([]map[string]interface{}, len(p.Skills))
		for i, s := range p.Skills {
			skills[i] = map[string]interface{}{
				"id":   s.Id,
				"name": s.Name,
			}
		}
		resp["skills"] = skills
	}

	return resp
}
