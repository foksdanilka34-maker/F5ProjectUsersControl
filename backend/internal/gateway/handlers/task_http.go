package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/business"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/gateway/websocket"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WebSocket event types
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type TaskHTTPHandler struct {
	client pb.BusinessServiceClient
	wsHub  *websocket.Hub
}

func NewTaskHTTPHandler(client pb.BusinessServiceClient, wsHub *websocket.Hub) *TaskHTTPHandler {
	return &TaskHTTPHandler{client: client, wsHub: wsHub}
}

// broadcastTaskEvent отправляет событие через WebSocket всем подключённым клиентам
func (h *TaskHTTPHandler) broadcastTaskEvent(eventType string, payload interface{}) {
	if h.wsHub == nil {
		log.Println("[WS] Hub is nil, skipping broadcast")
		return
	}
	log.Printf("[WS] Broadcasting event: %s to %d clients", eventType, h.wsHub.ClientCount())
	h.wsHub.BroadcastJSON(WSEvent{
		Type:    eventType,
		Payload: payload,
	})
}

// broadcastTaskEventFromProto - хелпер для отправки события с pb.Task
func (h *TaskHTTPHandler) broadcastTaskFromProto(eventType string, task *pb.Task) {
	h.broadcastTaskEvent(eventType, taskToMap(task))
}

func (h *TaskHTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tasks", h.Create)
	mux.HandleFunc("GET /api/v1/tasks", h.List)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.Get)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.Delete)
	mux.HandleFunc("POST /api/v1/tasks/{id}/move", h.Move)
	mux.HandleFunc("POST /api/v1/tasks/{id}/assign", h.Assign)
}

func (h *TaskHTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID   int64   `json:"project_id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		AssigneeID  *int64  `json:"assignee_id"`
		Priority    *int    `json:"priority"`
		DueDate     *string `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("task create decode error:", err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.ProjectID == 0 || req.Title == "" {
		http.Error(w, `{"error":"project_id and title required"}`, http.StatusBadRequest)
		return
	}

	// Получаем user_id из контекста (middleware должен положить)
	userID, _ := r.Context().Value("user_id").(int64)

	pbReq := &pb.CreateTaskRequest{
		ProjectId:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		CreatorId:   userID,
		AssigneeId:  req.AssigneeID,
	}
	if req.Priority != nil {
		priority := pb.TaskPriority(*req.Priority)
		pbReq.Priority = &priority
	}
	if req.DueDate != nil && *req.DueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			pbReq.DueDate = timestamppb.New(t)
		}
	}

	task, err := h.client.CreateTask(r.Context(), pbReq)
	if err != nil {
		log.Println("task create error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем событие через WebSocket
	h.broadcastTaskFromProto("task:created", task)

	writeJSON(w, http.StatusCreated, taskToMap(task))
}

func (h *TaskHTTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	task, err := h.client.GetTask(r.Context(), &pb.GetTaskRequest{TaskId: id})
	if err != nil {
		log.Println("task get error:", err)
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, taskToMap(task))
}

func (h *TaskHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	projectIDStr := query.Get("project_id")
	assigneeIDStr := query.Get("assignee_id")
	statusStr := query.Get("status")
	priorityStr := query.Get("priority")

	var projectID int64
	if projectIDStr != "" {
		projectID, _ = parseID(projectIDStr)
	}

	pbReq := &pb.ListTasksByProjectRequest{
		ProjectId: projectID,
	}
	if assigneeIDStr != "" {
		assigneeID, err := parseID(assigneeIDStr)
		if err == nil {
			pbReq.AssigneeId = &assigneeID
		}
	}
	if statusStr != "" {
		if statusVal, ok := pb.TaskStatus_value[statusStr]; ok {
			status := pb.TaskStatus(statusVal)
			pbReq.Status = &status
		}
	}
	if priorityStr != "" {
		if priorityVal, ok := pb.TaskPriority_value[priorityStr]; ok {
			priority := pb.TaskPriority(priorityVal)
			pbReq.Priority = &priority
		}
	}

	resp, err := h.client.ListTasksByProject(r.Context(), pbReq)
	if err != nil {
		log.Println("task list error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	tasks := make([]map[string]interface{}, len(resp.Tasks))
	for i, t := range resp.Tasks {
		tasks[i] = taskToMap(t)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

func (h *TaskHTTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		Priority    *int    `json:"priority"`
		AssigneeID  *int64  `json:"assignee_id"`
		DueDate     *string `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	pbReq := &pb.UpdateTaskRequest{
		TaskId:      id,
		Title:       req.Title,
		Description: req.Description,
		AssigneeId:  req.AssigneeID,
	}
	if req.Status != nil {
		if statusVal, ok := pb.TaskStatus_value[*req.Status]; ok {
			status := pb.TaskStatus(statusVal)
			pbReq.Status = &status
		}
	}
	if req.Priority != nil {
		priority := pb.TaskPriority(*req.Priority)
		pbReq.Priority = &priority
	}
	if req.DueDate != nil && *req.DueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			pbReq.DueDate = timestamppb.New(t)
		}
	}

	task, err := h.client.UpdateTask(r.Context(), pbReq)
	if err != nil {
		log.Println("task update error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем событие через WebSocket
	h.broadcastTaskFromProto("task:updated", task)

	writeJSON(w, http.StatusOK, taskToMap(task))
}

func (h *TaskHTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteTask(r.Context(), &pb.DeleteTaskRequest{TaskId: id})
	if err != nil {
		log.Println("task delete error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем событие через WebSocket
	h.broadcastTaskEvent("task:deleted", map[string]int64{"id": id})

	writeJSON(w, http.StatusOK, map[string]string{"message": "task deleted"})
}

func (h *TaskHTTPHandler) Move(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		NewStatus     string `json:"new_status"`
		NewOrderIndex int    `json:"new_order_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewStatus == "" {
		http.Error(w, `{"error":"new_status required"}`, http.StatusBadRequest)
		return
	}

	statusVal, ok := pb.TaskStatus_value[req.NewStatus]
	if !ok {
		http.Error(w, `{"error":"invalid status"}`, http.StatusBadRequest)
		return
	}

	task, err := h.client.MoveTask(r.Context(), &pb.MoveTaskRequest{
		TaskId:        id,
		NewStatus:     pb.TaskStatus(statusVal),
		NewOrderIndex: int32(req.NewOrderIndex),
	})
	if err != nil {
		log.Println("task move error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем событие через WebSocket
	h.broadcastTaskFromProto("task:moved", task)

	writeJSON(w, http.StatusOK, taskToMap(task))
}

func (h *TaskHTTPHandler) Assign(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		AssigneeID int64 `json:"assignee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AssigneeID == 0 {
		http.Error(w, `{"error":"assignee_id required"}`, http.StatusBadRequest)
		return
	}

	task, err := h.client.AssignTask(r.Context(), &pb.AssignTaskRequest{
		TaskId:     id,
		AssigneeId: req.AssigneeID,
	})
	if err != nil {
		log.Println("task assign error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем событие через WebSocket
	h.broadcastTaskFromProto("task:assigned", task)

	writeJSON(w, http.StatusOK, taskToMap(task))
}

func taskToMap(t *pb.Task) map[string]interface{} {
	resp := map[string]interface{}{
		"id":          t.Id,
		"project_id":  t.ProjectId,
		"title":       t.Title,
		"description": t.Description,
		"status":      t.Status.String(),
		"priority":    t.Priority.String(),
		"assignee_id": t.AssigneeId,
		"creator_id":  t.CreatorId,
		"order_index": t.OrderIndex,
	}
	if t.DueDate != nil {
		resp["due_date"] = t.DueDate.AsTime()
	}
	if t.StartedAt != nil {
		resp["started_at"] = t.StartedAt.AsTime()
	}
	if t.CompletedAt != nil {
		resp["completed_at"] = t.CompletedAt.AsTime()
	}
	if t.CreatedAt != nil {
		resp["created_at"] = t.CreatedAt.AsTime()
	}
	if t.UpdatedAt != nil {
		resp["updated_at"] = t.UpdatedAt.AsTime()
	}
	return resp
}
