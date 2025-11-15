package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenHeader     = "X-Refresh-Token"
)

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
