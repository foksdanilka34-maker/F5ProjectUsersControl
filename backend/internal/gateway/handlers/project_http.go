package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/business"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectHTTPHandler struct {
	client pb.BusinessServiceClient
}

func NewProjectHTTPHandler(client pb.BusinessServiceClient) *ProjectHTTPHandler {
	return &ProjectHTTPHandler{client: client}
}

func (h *ProjectHTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/projects", h.Create)
	mux.HandleFunc("GET /api/v1/projects", h.List)
	mux.HandleFunc("GET /api/v1/projects/{id}", h.Get)
	mux.HandleFunc("PUT /api/v1/projects/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", h.Delete)
	mux.HandleFunc("POST /api/v1/projects/{id}/members", h.AddMember)
	mux.HandleFunc("DELETE /api/v1/projects/{id}/members/{userId}", h.RemoveMember)
	mux.HandleFunc("GET /api/v1/projects/{id}/members", h.ListMembers)
}

func (h *ProjectHTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		ManagerID   int64   `json:"manager_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("project create decode error:", err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.ManagerID == 0 {
		http.Error(w, `{"error":"name and manager_id required"}`, http.StatusBadRequest)
		return
	}

	pbReq := &pb.CreateProjectRequest{
		Name:      req.Name,
		ManagerId: req.ManagerID,
	}
	if req.Description != nil {
		pbReq.Description = req.Description
	}

	project, err := h.client.CreateProject(r.Context(), pbReq)
	if err != nil {
		log.Println("project create error:", err)
		errMsg := err.Error()
		// Проверяем на ошибку уникальности
		if strings.Contains(errMsg, "already exists") {
			http.Error(w, `{"error":"Проект с таким именем уже существует"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"`+errMsg+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, projectToMap(project))
}

func (h *ProjectHTTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	project, err := h.client.GetProject(r.Context(), &pb.GetProjectRequest{ProjectId: id})
	if err != nil {
		log.Println("project get error:", err)
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, projectToMap(project))
}

func (h *ProjectHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	pageNumber, _ := strconv.Atoi(query.Get("page_number"))
	if pageSize == 0 {
		pageSize = 20
	}
	if pageNumber == 0 {
		pageNumber = 1
	}

	pbReq := &pb.ListProjectsRequest{
		PageSize:   int32(pageSize),
		PageNumber: int32(pageNumber),
	}

	if managerIDStr := query.Get("manager_id"); managerIDStr != "" {
		managerID, err := parseID(managerIDStr)
		if err == nil {
			pbReq.ManagerId = &managerID
		}
	}
	if memberIDStr := query.Get("member_id"); memberIDStr != "" {
		memberID, err := parseID(memberIDStr)
		if err == nil {
			pbReq.MemberId = &memberID
		}
	}
	if statusStr := query.Get("status"); statusStr != "" {
		if statusVal, ok := pb.ProjectStatus_value[statusStr]; ok {
			status := pb.ProjectStatus(statusVal)
			pbReq.Status = &status
		}
	}

	resp, err := h.client.ListProjects(r.Context(), pbReq)
	if err != nil {
		// Don't log context canceled errors (normal when client disconnects)
		if r.Context().Err() != context.Canceled && status.Code(err) != codes.Canceled {
			log.Println("project list error:", err)
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	projects := make([]map[string]interface{}, len(resp.Projects))
	for i, p := range resp.Projects {
		projects[i] = projectToMap(p)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"projects":    projects,
		"total_count": resp.TotalCount,
	})
}

func (h *ProjectHTTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	pbReq := &pb.UpdateProjectRequest{
		ProjectId:   id,
		Name:        req.Name,
		Description: req.Description,
	}
	if req.Status != nil {
		if statusVal, ok := pb.ProjectStatus_value[*req.Status]; ok {
			status := pb.ProjectStatus(statusVal)
			pbReq.Status = &status
		}
	}

	project, err := h.client.UpdateProject(r.Context(), pbReq)
	if err != nil {
		log.Println("project update error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, projectToMap(project))
}

func (h *ProjectHTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteProject(r.Context(), &pb.DeleteProjectRequest{ProjectId: id})
	if err != nil {
		log.Println("project delete error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "project deleted"})
}

func (h *ProjectHTTPHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("id")
	projectID, err := parseID(projectIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid project_id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 {
		http.Error(w, `{"error":"user_id required"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.AddMemberToProject(r.Context(), &pb.AddMemberToProjectRequest{
		ProjectId: projectID,
		UserId:    req.UserID,
	})
	if err != nil {
		log.Println("add member error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member added"})
}

func (h *ProjectHTTPHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("id")
	projectID, err := parseID(projectIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid project_id"}`, http.StatusBadRequest)
		return
	}
	userIDStr := r.PathValue("userId")
	userID, err := parseID(userIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid user_id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.RemoveMemberFromProject(r.Context(), &pb.RemoveMemberFromProjectRequest{
		ProjectId: projectID,
		UserId:    userID,
	})
	if err != nil {
		log.Println("remove member error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

func (h *ProjectHTTPHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("id")
	projectID, err := parseID(projectIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid project_id"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.ListProjectMembers(r.Context(), &pb.ListProjectMembersRequest{ProjectId: projectID})
	if err != nil {
		log.Println("list members error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	members := make([]map[string]interface{}, len(resp.Members))
	for i, m := range resp.Members {
		members[i] = map[string]interface{}{
			"user_id":   m.UserId,
			"full_name": m.FullName,
			"role":      m.Role,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"members": members})
}

func projectToMap(p *pb.Project) map[string]interface{} {
	resp := map[string]interface{}{
		"id":          p.Id,
		"name":        p.Name,
		"description": p.Description,
		"manager_id":  p.ManagerId,
		"status":      p.Status.String(),
	}
	if p.DueDate != nil {
		resp["due_date"] = p.DueDate.AsTime()
	}
	if p.CreatedAt != nil {
		resp["created_at"] = p.CreatedAt.AsTime()
	}
	if p.UpdatedAt != nil {
		resp["updated_at"] = p.UpdatedAt.AsTime()
	}
	return resp
}
