// main.go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin_saas_auth/internal/api/middleware"
	v1 "gin_saas_auth/internal/api/v1"
	"gin_saas_auth/internal/config"
	"gin_saas_auth/internal/services"

	_ "gin_saas_auth/api/swagger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title 认证服务API
// @version 1.0
// @description 基于Gin框架的SaaS认证服务
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
func main() {
	// 初始化日志
	middleware.InitLogger()
	logrus.Info("开始初始化认证服务...")

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("加载配置失败: %v", err)
	}
	logrus.Info("配置加载成功")

	// 设置Gin模式
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		logrus.Info("运行在生产模式")
	} else {
		gin.SetMode(gin.DebugMode)
		logrus.Info("运行在开发模式")
	}

	// 设置路由
	r := v1.SetupRouter(cfg.Server.Domain)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 初始化 Consul 注册（如果启用）
	var consulRegistry *services.ConsulRegistry
	if cfg.IsConsulEnabled() {
		logrus.Info("正在初始化 Consul 服务注册...")

		// 验证 Consul 配置
		if err := services.ValidateConfig(cfg); err != nil {
			logrus.Warnf("Consul 配置验证失败: %v", err)
		} else {
			consulRegistry, err = services.NewConsulRegistry(cfg)
			if err != nil {
				logrus.Errorf("创建 Consul 注册器失败: %v", err)
			} else {
				// 注册服务到 Consul
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := consulRegistry.Register(ctx); err != nil {
					logrus.Errorf("注册服务到 Consul 失败: %v", err)
				}
				cancel()
			}
		}
	} else {
		logrus.Info("Consul 服务发现未启用")
	}

	// 启动服务器（非阻塞）
	go func() {
		logrus.WithFields(logrus.Fields{
			"host":    cfg.Server.Host,
			"port":    cfg.App.Port,
			"domain":  cfg.Server.Domain,
			"env":     cfg.App.Env,
			"version": cfg.Consul.Meta.Version,
		}).Info("OSS文件转发服务启动成功")

		logrus.Infof("外部访问地址: %s", cfg.GetServiceURL())
		logrus.Infof("Swagger文档地址: %s/swagger/index.html", cfg.GetServiceURL())
		logrus.Infof("健康检查地址: %s/health", cfg.GetServiceURL())
		logrus.Infof("监控指标地址: %s/metrics", cfg.GetServiceURL())
		logrus.Infof("服务统计地址: %s/api/v1/stats", cfg.GetServiceURL())

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务器...")

	// 设置5秒的超时时间来关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 从 Consul 注销服务
	if consulRegistry != nil {
		logrus.Info("正在从 Consul 注销服务...")
		if err := consulRegistry.Deregister(ctx); err != nil {
			logrus.Errorf("从 Consul 注销服务失败: %v", err)
		} else {
			logrus.Info("服务已从 Consul 注销")
		}
	}

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("服务器强制关闭: %v", err)
	}

	logrus.Info("OSS文件转发服务已安全关闭")
}
