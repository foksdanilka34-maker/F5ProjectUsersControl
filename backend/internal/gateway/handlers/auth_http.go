package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/identity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthHTTPHandler struct {
	client pb.IdentityServiceClient
}

func NewAuthHTTPHandler(client pb.IdentityServiceClient) *AuthHTTPHandler {
	return &AuthHTTPHandler{client: client}
}

func (h *AuthHTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.RefreshToken)
	mux.HandleFunc("GET /api/v1/auth/me", h.GetMe)
	mux.HandleFunc("POST /api/v1/auth/change-password", h.ChangePassword)
}

func (h *AuthHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("auth login decode error:", err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, `{"error":"login and password required"}`, http.StatusBadRequest)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	clientIP := getClientIP(r)

	ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs(
		"user-agent", userAgent,
		"x-forwarded-for", clientIP,
	))

	resp, err := h.client.Login(ctx, &pb.LoginRequest{
		Login:     req.Login,
		Password:  req.Password,
		UserAgent: userAgent,
		IpAddress: clientIP,
	})
	if err != nil {
		log.Println("auth login error:", err)
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": resp.AccessToken,
		"user": map[string]interface{}{
			"id":         resp.User.Id,
			"login":      resp.User.Login,
			"full_name":  resp.User.FullName,
			"role":       resp.User.Role,
			"avatar_url": resp.User.AvatarUrl,
		},
	})
}

func (h *AuthHTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs(
		"authorization", r.Header.Get("Authorization"),
	))

	_, err := h.client.Logout(ctx, &emptypb.Empty{})
	if err != nil {
		log.Println("auth logout error:", err)
		http.Error(w, `{"error":"logout failed"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *AuthHTTPHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs(
		"authorization", r.Header.Get("Authorization"),
		"user-agent", r.Header.Get("User-Agent"),
		"x-forwarded-for", getClientIP(r),
	))

	resp, err := h.client.Refresh(ctx, &emptypb.Empty{})
	if err != nil {
		log.Println("auth refresh error:", err)
		http.Error(w, `{"error":"refresh failed"}`, http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"access_token": resp.AccessToken})
}

func (h *AuthHTTPHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// user_id должен быть в контексте от middleware
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok || userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs(
		"x-user-id", fmt.Sprintf("%d", userID),
	))

	resp, err := h.client.GetMe(ctx, &emptypb.Empty{})
	if err != nil {
		log.Println("get me error:", err)
		http.Error(w, `{"error":"failed to get user info"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         resp.Id,
		"login":      resp.Login,
		"full_name":  resp.FullName,
		"role":       resp.Role,
		"avatar_url": resp.AvatarUrl,
	})
}

func (h *AuthHTTPHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      int64  `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || len(req.NewPassword) < 6 {
		http.Error(w, `{"error":"user_id required and password min 6 chars"}`, http.StatusBadRequest)
		return
	}

	_, err := h.client.ChangePassword(r.Context(), &pb.ChangePasswordRequest{
		UserId:      req.UserID,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		log.Println("change password error:", err)
		http.Error(w, `{"error":"failed to change password"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}
