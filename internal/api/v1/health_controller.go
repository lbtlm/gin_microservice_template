package v1

import (
	"fmt"
	"net/http"
	"time"

	"gin_saas_auth/internal/config"
	"gin_saas_auth/internal/services"

	"github.com/gin-gonic/gin"
)

// PingHandler 测试接口
func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "pong",
		"status":    "success",
		"timestamp": time.Now().Unix(),
	})
}

// HealthHandler 健康检查接口
func HealthHandler(c *gin.Context) {
	cfg := config.GlobalConfig

	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"message":   "认证服务运行正常",
		"timestamp": time.Now().Unix(),
		"service": map[string]interface{}{
			"name":    cfg.App.Name,
			"version": cfg.Consul.Meta.Version,
			"env":     cfg.App.Env,
		},
		"dependencies": map[string]interface{}{},
	}

	// 如果启用了 Consul，检查 Consul 连接状态
	if cfg.IsConsulEnabled() {
		consulRegistry, err := services.NewConsulRegistry(cfg)
		if err != nil {
			healthStatus["dependencies"].(map[string]interface{})["consul"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			consulHealthy := consulRegistry.IsHealthy(c.Request.Context())
			consulStatus := "healthy"
			if !consulHealthy {
				consulStatus = "unhealthy"
			}
			healthStatus["dependencies"].(map[string]interface{})["consul"] = map[string]interface{}{
				"status":  consulStatus,
				"address": cfg.Consul.Address,
			}
		}
	}

	c.JSON(http.StatusOK, healthStatus)
}

// MetricsHandler 监控指标接口
func MetricsHandler(c *gin.Context) {
	cfg := config.GlobalConfig
	now := time.Now().Unix()

	metricsText := fmt.Sprintf(`# HELP auth_service_info 认证服务信息
# TYPE auth_service_info gauge
auth_service_info{name="%s",version="%s",framework="%s",language="%s",environment="%s"} 1

# HELP service_uptime_seconds 服务启动时间（秒）
# TYPE service_uptime_seconds gauge
service_uptime_seconds %d
`, cfg.App.Name, cfg.Consul.Meta.Version, cfg.Consul.Meta.Framework, cfg.Consul.Meta.Language, cfg.App.Env, now)

	c.String(http.StatusOK, metricsText)
}

// StatsHandler 服务统计信息接口
func StatsHandler(c *gin.Context) {
	cfg := config.GlobalConfig
	c.JSON(http.StatusOK, gin.H{
		"service_name": cfg.App.Name,
		"version":      cfg.Consul.Meta.Version,
		"environment":  cfg.App.Env,
		"start_time":   time.Now().Unix(),
	})
}

// ConsulInfoHandler Consul 配置信息接口
func ConsulInfoHandler(c *gin.Context) {
	cfg := config.GlobalConfig
	c.JSON(http.StatusOK, cfg.Consul)
}
