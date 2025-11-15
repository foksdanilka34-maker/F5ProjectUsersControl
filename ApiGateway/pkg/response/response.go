package response

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type PaginationMeta struct {
	PageSize   int32 `json:"page_size,omitempty"`
	PageNumber int32 `json:"page_number,omitempty"`
	TotalCount int64 `json:"total_count,omitempty"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

func Success(c *gin.Context, statusCode int, data interface{}, message string) {
	write(c, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta interface{}, message string) {
	write(c, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Error(c *gin.Context, statusCode int, errMessage string) {
	write(c, statusCode, Response{
		Success: false,
		Error:   errMessage,
	})
}

func Paginated(c *gin.Context, statusCode int, data interface{}, meta PaginationMeta, message string) {
	SuccessWithMeta(c, statusCode, data, meta, message)
}

func Created(c *gin.Context, data any, message string) {
	Success(c, 201, data, message)
}

func BadRequest(c *gin.Context, errMessage string) {
	Error(c, 400, errMessage)
}

func Unauthorized(c *gin.Context, errMessage string) {
	if errMessage == "" {
		errMessage = "Unauthorized"
	}
	Error(c, 401, errMessage)
}

func Forbidden(c *gin.Context, errMessage string) {
	if errMessage == "" {
		errMessage = "Access forbidden"
	}
	Error(c, 403, errMessage)
}

func NotFound(c *gin.Context, errMessage string) {
	if errMessage == "" {
		errMessage = "Resource not found"
	}
	Error(c, 404, errMessage)
}

func InternalServerError(c *gin.Context, errMessage string) {
	if errMessage == "" {
		errMessage = "Internal server error"
	}
	Error(c, 500, errMessage)
}

func write(c *gin.Context, statusCode int, payload Response) {
	c.JSON(statusCode, payload)
}
