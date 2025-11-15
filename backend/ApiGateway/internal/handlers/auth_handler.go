package handlers

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/config"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/models"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName   = "refresh_token"
	refreshTokenHeader       = "X-Refresh-Token"
	refreshTokenCookieMaxAge = 30 * 24 * 60 * 60
)

type AuthHandler struct {
	loginService *service.LoginServiceClient
	config       *config.Config
}

func NewAuthHandler(loginService *service.LoginServiceClient, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		loginService: loginService,
		config:       cfg,
	}
}

// Login authenticates a user and issues access and refresh tokens.
// @Summary      Authenticate user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        login body models.LoginRequest true "Login credentials"
// @Success      200  {object} response.Response{data=response.LoginResponse}
// @Failure      400  {object} response.Response
// @Failure      401  {object} response.Response
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Login validation error: %v", err)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	log.Printf("Login attempt: login=%s, userAgent=%s, ip=%s", req.Login, userAgent, clientIP)

	accessToken, refreshToken, err := h.loginService.Login(c, req.Login, req.Password, userAgent, clientIP)
	if err != nil {
		log.Printf("Login service error: %v", err)
		response.Unauthorized(c, "Invalid login or password")
		return
	}

	log.Printf("Login successful for user: %s", req.Login)

	h.setRefreshTokenCookie(c, refreshToken)

	data := response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60,
	}

	response.Success(c, http.StatusOK, data, "Login successful")
}

// Logout revokes the refresh token stored in cookies or headers.
// @Summary      Logout current user
// @Tags         Auth
// @Security     ApiKeyAuth
// @Produce      json
// @Param        Authorization header string false "Bearer token"
// @Param        X-Refresh-Token header string false "Refresh token"
// @Success      200  {object} response.Response
// @Failure      400  {object} response.Response
// @Failure      401  {object} response.Response
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	log.Printf("Logout attempt: userID=%s", userID)

	refreshToken := getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		log.Printf("Logout: missing refresh token")
		response.Unauthorized(c, "Missing refresh token")
		return
	}

	err := h.loginService.Logout(c, refreshToken)
	if err != nil {
		log.Printf("Logout service error: %v", err)
	}

	h.clearRefreshTokenCookie(c)

	log.Printf("Logout successful for user: %s", userID)

	response.Success(c, http.StatusOK, nil, "Logout successful")
}

// Refresh returns a new access token when provided with a valid refresh token.
// @Summary      Refresh access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        refresh_token body models.RefreshTokenRequest false "Refresh token"
// @Param        Authorization header string false "Bearer token"
// @Param        X-Refresh-Token header string false "Refresh token"
// @Success      200  {object} response.Response{data=response.RefreshResponse}
// @Failure      400  {object} response.Response
// @Failure      401  {object} response.Response
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	log.Printf("Refresh attempt")

	refreshToken := getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		var body models.RefreshTokenRequest
		if err := c.ShouldBindJSON(&body); err == nil {
			refreshToken = strings.TrimSpace(body.RefreshToken)
		}
	}
	if refreshToken == "" {
		refreshToken = extractBearerToken(c.GetHeader("Authorization"))
	}
	if refreshToken == "" {
		log.Printf("Refresh: missing refresh token")
		response.Unauthorized(c, "Missing refresh token")
		return
	}

	decodedRefreshToken, err := url.QueryUnescape(refreshToken)
	if err == nil {
		refreshToken = decodedRefreshToken
	}

	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	newAccessToken, newRefreshToken, err := h.loginService.Refresh(c, refreshToken, userAgent, clientIP)
	if err != nil {
		log.Printf("Refresh service error: %v", err)
		response.Unauthorized(c, "Token refresh failed")
		return
	}

	h.setRefreshTokenCookie(c, newRefreshToken)

	log.Printf("Refresh successful")

	data := response.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    15 * 60,
	}

	response.Success(c, http.StatusOK, data, "Token refreshed successfully")
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	if token == "" {
		return
	}

	isSecure := h.config.Environment == "production"
	c.SetCookie(refreshTokenCookieName, token, refreshTokenCookieMaxAge, "/", "", isSecure, true)
}

func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	isSecure := h.config.Environment == "production"
	c.SetCookie(refreshTokenCookieName, "", -1, "/", "", isSecure, true)
}

func getRefreshTokenFromRequest(c *gin.Context) string {
	if token, err := c.Cookie(refreshTokenCookieName); err == nil && token != "" {
		return token
	}

	header := c.GetHeader(refreshTokenHeader)
	if header != "" {
		return strings.TrimSpace(header)
	}

	return ""
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
