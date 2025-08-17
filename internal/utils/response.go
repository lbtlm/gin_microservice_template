// utils/response.go
package utils

import (
	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// SuccessWithData 成功响应带数据
func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: "操作成功",
		Data:    data,
	})
}

// Success 成功响应
func Success(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    200,
		Message: message,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	ErrorResponse(c, 400, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, 401, message)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	ErrorResponse(c, 403, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	ErrorResponse(c, 404, message)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string) {
	ErrorResponse(c, 500, message)
}
