package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/http/middleware"
)

type TaskService interface {
	CreateTask(ctx context.Context, creatorID int64, req dto.CreateTaskRequest) (dto.TaskDTO, error)
	GetTask(ctx context.Context, id int64) (dto.TaskDTO, error)
	ListTasks(ctx context.Context, filter dto.ListTasksFilter) ([]dto.TaskDTO, error)
	UpdateTask(ctx context.Context, id, userID int64, req dto.UpdateTaskRequest) (dto.TaskDTO, error)
	DeleteTask(ctx context.Context, id int64) error
	MoveTask(ctx context.Context, id int64, req dto.MoveTaskRequest) (dto.TaskDTO, error)
	AssignTask(ctx context.Context, id, assigneeID int64) (dto.TaskDTO, error)

	SubmitForReview(ctx context.Context, id int64) (dto.TaskDTO, error)
	ApproveTask(ctx context.Context, id, userID int64) (dto.TaskDTO, error)
	GetReviewStatus(ctx context.Context, id int64) (dto.ReviewStatusResponse, error)

	AddComment(ctx context.Context, taskID, userID int64, content string) (dto.TaskCommentDTO, error)
	ListComments(ctx context.Context, taskID int64) ([]dto.TaskCommentDTO, error)
}

type TaskHandler struct {
	service       TaskService
	authValidator *middleware.JWTValidator
}

func NewTaskHandler(mux *http.ServeMux, service TaskService, authValidator *middleware.JWTValidator) *TaskHandler {
	h := &TaskHandler{
		service:       service,
		authValidator: authValidator,
	}
	h.registerRoutes(mux)
	return h
}

func (h *TaskHandler) registerRoutes(mux *http.ServeMux) {
	authMW := middleware.Auth(h.authValidator)

	mux.Handle("GET /api/v1/tasks", authMW(middleware.Chaos(http.HandlerFunc(h.ListTasks))))
	mux.Handle("GET /api/v1/tasks/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.GetTask))))
	mux.Handle("POST /api/v1/tasks", authMW(middleware.Chaos(http.HandlerFunc(h.CreateTask))))
	mux.Handle("PUT /api/v1/tasks/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.UpdateTask))))
	mux.Handle("DELETE /api/v1/tasks/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.DeleteTask))))

	mux.Handle("POST /api/v1/tasks/{id}/move", authMW(middleware.Chaos(http.HandlerFunc(h.MoveTask))))
	mux.Handle("POST /api/v1/tasks/{id}/assign", authMW(middleware.Chaos(http.HandlerFunc(h.AssignTask))))
	mux.Handle("POST /api/v1/tasks/{id}/review/submit", authMW(middleware.Chaos(http.HandlerFunc(h.SubmitForReview))))
	mux.Handle("POST /api/v1/tasks/{id}/review/approve", authMW(middleware.Chaos(http.HandlerFunc(h.ApproveTask))))
	mux.Handle("GET /api/v1/tasks/{id}/review", authMW(middleware.Chaos(http.HandlerFunc(h.GetReviewStatus))))

	mux.Handle("GET /api/v1/tasks/{id}/comments", authMW(middleware.Chaos(http.HandlerFunc(h.ListComments))))
	mux.Handle("POST /api/v1/tasks/{id}/comments", authMW(middleware.Chaos(http.HandlerFunc(h.AddComment))))
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	projectID, _ := strconv.ParseInt(query.Get("project_id"), 10, 64)
	assigneeID, _ := strconv.ParseInt(query.Get("assignee_id"), 10, 64)
	status := query.Get("status")
	priority := query.Get("priority")

	filter := dto.ListTasksFilter{
		ProjectID:  projectID,
		AssigneeID: assigneeID,
		Status:     status,
		Priority:   priority,
	}

	tasks, err := h.service.ListTasks(r.Context(), filter)
	if err != nil {
		log.Println("list tasks error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"tasks": tasks})
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	task, err := h.service.GetTask(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("create task decode error:", err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(r.Context(), userID, req)
	if err != nil {
		log.Println("create task error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	task, err := h.service.UpdateTask(r.Context(), id, userID, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "task deleted"})
}

func (h *TaskHandler) MoveTask(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	var req dto.MoveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewStatus == "" {
		http.Error(w, `{"error":"new_status is required"}`, http.StatusBadRequest)
		return
	}

	task, err := h.service.MoveTask(r.Context(), id, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		AssigneeID int64 `json:"assignee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AssigneeID == 0 {
		http.Error(w, `{"error":"assignee_id is required"}`, http.StatusBadRequest)
		return
	}

	task, err := h.service.AssignTask(r.Context(), id, req.AssigneeID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	task, err := h.service.SubmitForReview(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) ApproveTask(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)
	if userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	task, err := h.service.ApproveTask(r.Context(), id, userID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetReviewStatus(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	res, err := h.service.GetReviewStatus(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *TaskHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	comments, err := h.service.ListComments(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"comments": comments})
}

func (h *TaskHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)
	if userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	comment, err := h.service.AddComment(r.Context(), id, userID, req.Content)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(comment)
}
