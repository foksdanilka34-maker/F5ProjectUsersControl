package middleware

import (
	"log"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	"github.com/gin-gonic/gin"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				response.InternalServerError(c, "Internal server error")
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Handler error: %v", err)
		}
	}
}
