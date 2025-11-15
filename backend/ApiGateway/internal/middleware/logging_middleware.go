package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		log.Printf(
			"[%s] %s %s | Status: %d | Duration: %v | Client: %s",
			method,
			path,
			c.Request.Proto,
			statusCode,
			duration,
			clientIP,
		)
	}
}
