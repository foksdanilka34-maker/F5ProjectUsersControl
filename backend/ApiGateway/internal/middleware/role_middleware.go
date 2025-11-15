package middleware

import (
	"log"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	"github.com/gin-gonic/gin"
)

func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetRoleFromContext(c)
		if userRole == "" {
			log.Println("User role not found in context")
			response.Unauthorized(c, "User role not found")
			c.Abort()
			return
		}

		hasRequiredRole := false
		for _, required := range requiredRoles {
			if userRole == required {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			userID := GetUserIDFromContext(c)
			log.Printf("Access denied for user %s with role %s", userID, userRole)
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		log.Printf("User with role %s has access", userRole)
		c.Next()
	}
}
