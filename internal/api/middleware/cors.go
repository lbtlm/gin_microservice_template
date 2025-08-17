// middleware/cors.go
package middleware

import (
	"gin_saas_auth/internal/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			// 获取配置的允许源
			allowedOrigins := config.GlobalConfig.Server.AllowedOrigins

			// 检查是否允许该源
			var allowOrigin string
			if allowedOrigins == "*" {
				// 允许所有源
				allowOrigin = "*"
			} else {
				// 检查源是否在允许列表中
				origins := strings.Split(allowedOrigins, ",")
				for _, ao := range origins {
					ao = strings.TrimSpace(ao)
					if ao == origin {
						allowOrigin = origin
						break
					}
				}
				// 如果未找到匹配的源，则不设置允许源头
				if allowOrigin == "" {
					allowOrigin = "null" // 明确拒绝
				}
			}

			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Token, X-Token")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
