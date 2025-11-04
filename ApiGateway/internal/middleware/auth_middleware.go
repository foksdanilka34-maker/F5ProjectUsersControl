package middleware

import (
	"log"
	"strings"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

const (
	UserIDCtxKey = "user_id"
	RoleCtxKey   = "role"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Missing authorization header")
			response.Unauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Println("Invalid authorization header format")
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			log.Println("Empty token")
			response.Unauthorized(c, "Empty token")
			c.Abort()
			return
		}

		claims := &CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set(UserIDCtxKey, claims.UserID)
		c.Set(RoleCtxKey, claims.Role)

		log.Printf("User authenticated: userID=%s, role=%s", claims.UserID, claims.Role)

		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) string {
	userID, exists := c.Get(UserIDCtxKey)
	if !exists {
		return ""
	}
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

func GetRoleFromContext(c *gin.Context) string {
	role, exists := c.Get(RoleCtxKey)
	if !exists {
		return ""
	}
	if r, ok := role.(string); ok {
		return r
	}
	return ""
}
