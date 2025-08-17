// internal/api/v1/routes.go
package v1

import (
	"gin_saas_auth/internal/api/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter(domain string) *gin.Engine {
	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// 添加Swagger文档路由 - 动态配置host
	url := ginSwagger.URL("http://" + domain + "/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// 健康检查和监控接口
	r.GET("/health", HealthHandler)
	r.GET("/ping", PingHandler)
	r.GET("/metrics", MetricsHandler)

	// API路由分组
	api := r.Group("/api")
	{
		// v1 版本分组
		v1Group := api.Group("/v1")
		{
			// 服务统计信息
			v1Group.GET("/stats", StatsHandler)

			// Consul 相关接口
			consulGroup := v1Group.Group("/consul")
			{
				consulGroup.GET("/info", ConsulInfoHandler)
			}
		}
	}

	return r
}
