package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/identity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenMaxAge     = 7 * 24 * 60 * 60 // 7 days in seconds
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

	// Set refresh token as HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    resp.RefreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   refreshTokenMaxAge,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

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
	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Get refresh token from cookie to invalidate session
	cookie, _ := r.Cookie(refreshTokenCookieName)
	if cookie != nil && cookie.Value != "" {
		ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs(
			"authorization", cookie.Value,
		))
		_, _ = h.client.Logout(ctx, &emptypb.Empty{})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *AuthHTTPHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Read refresh token from HttpOnly cookie
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil || cookie.Value == "" {
		http.Error(w, `{"error":"refresh token not found"}`, http.StatusUnauthorized)
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs(
		"authorization", cookie.Value,
		"user-agent", r.Header.Get("User-Agent"),
		"x-forwarded-for", getClientIP(r),
	))

	resp, err := h.client.Refresh(ctx, &emptypb.Empty{})
	if err != nil {
		// Don't log context canceled errors (normal when client disconnects)
		if r.Context().Err() != context.Canceled && status.Code(err) != codes.Canceled {
			log.Println("auth refresh error:", err)
		}
		// Clear invalid cookie
		http.SetCookie(w, &http.Cookie{
			Name:     refreshTokenCookieName,
			Value:    "",
			Path:     "/api/v1/auth",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
		http.Error(w, `{"error":"refresh failed"}`, http.StatusUnauthorized)
		return
	}

	// Set new refresh token cookie (token rotation)
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    resp.RefreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   refreshTokenMaxAge,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Return access_token and user info
	result := map[string]interface{}{
		"access_token": resp.AccessToken,
	}
	if resp.User != nil {
		result["user"] = map[string]interface{}{
			"id":         resp.User.Id,
			"login":      resp.User.Login,
			"full_name":  resp.User.FullName,
			"role":       resp.User.Role,
			"avatar_url": resp.User.AvatarUrl,
		}
	}
	writeJSON(w, http.StatusOK, result)
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
