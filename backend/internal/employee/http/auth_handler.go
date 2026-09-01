package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/http/middleware"
)

const (
	refreshTokenCookie = "refresh_token"
	refreshTokenMaxAge = 7 * 24 * 60 * 60
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken, userAgent, ipAddress string) (dto.RefreshResponse, error)
	GetMe(ctx context.Context, userID int64) (dto.UserInfo, error)
	ChangePassword(ctx context.Context, userID int64, newPassword string) error
	ValidateToken(tokenStr string) (int64, string, error)
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(mux *http.ServeMux, service AuthService) *AuthHandler {
	h := &AuthHandler{service: service}
	h.registerRoutes(mux)
	return h
}

func (h *AuthHandler) registerRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/auth/login", middleware.Chaos(http.HandlerFunc(h.Login)))
	mux.Handle("POST /api/v1/auth/logout", middleware.Chaos(http.HandlerFunc(h.Logout)))
	mux.Handle("POST /api/v1/auth/refresh", middleware.Chaos(http.HandlerFunc(h.Refresh)))

	// Protected routes
	authMW := middleware.Auth(h.service)
	mux.Handle("GET /api/v1/auth/me", authMW(middleware.Chaos(http.HandlerFunc(h.GetMe))))
	mux.Handle("POST /api/v1/auth/change-password", authMW(middleware.Chaos(http.HandlerFunc(h.ChangePassword))))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("error decoding login json:", err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	req.UserAgent = r.Header.Get("User-Agent")
	req.IPAddress = r.RemoteAddr

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		log.Println("login failed:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    resp.RefreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   refreshTokenMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var refreshToken string
	if cookie, err := r.Cookie(refreshTokenCookie); err == nil {
		refreshToken = cookie.Value
	}

	_ = h.service.Logout(r.Context(), refreshToken)

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string
	if cookie, err := r.Cookie(refreshTokenCookie); err == nil {
		refreshToken = cookie.Value
	}
	if refreshToken == "" {
		http.Error(w, `{"error":"missing refresh token cookie"}`, http.StatusUnauthorized)
		return
	}

	resp, err := h.service.Refresh(r.Context(), refreshToken, r.Header.Get("User-Agent"), r.RemoteAddr)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     refreshTokenCookie,
			Value:    "",
			Path:     "/api/v1/auth",
			MaxAge:   -1,
			HttpOnly: true,
		})
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    resp.RefreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   refreshTokenMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok || userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok || userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	targetID := req.UserID
	if targetID == 0 {
		targetID = userID
	}

	if err := h.service.ChangePassword(r.Context(), targetID, req.NewPassword); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "password changed successfully"})
}
