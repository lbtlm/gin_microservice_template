// middleware/logger.go
package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	// FileLogger 文件日志实例
	FileLogger *logrus.Logger
	// HTTPLogger HTTP请求日志实例
	HTTPLogger *logrus.Logger
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用HTTP专用日志记录器
		HTTPLogger.WithFields(logrus.Fields{
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
		}).Info("HTTP Request")

		// 返回空字符串，因为我们已经通过logrus记录了日志
		return ""
	})
}

// InitLogger 初始化日志系统
func InitLogger() {
	// 基础日志配置
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 创建文件日志实例
	FileLogger = logrus.New()
	FileLogger.SetLevel(logrus.InfoLevel)
	FileLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 创建HTTP日志实例
	HTTPLogger = logrus.New()
	HTTPLogger.SetLevel(logrus.InfoLevel)
	HTTPLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 确保日志目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		logrus.Warn("创建日志目录失败:", err)
		return
	}

	// 设置文件输出
	setupFileOutput()
}

// setupFileOutput 设置文件输出
func setupFileOutput() {
	// 应用日志文件
	appLogFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Warn("无法打开应用日志文件:", err)
	} else {
		FileLogger.SetOutput(io.MultiWriter(os.Stdout, appLogFile))
		logrus.SetOutput(io.MultiWriter(os.Stdout, appLogFile))
	}

	// HTTP请求日志文件
	httpLogFile, err := os.OpenFile("logs/http.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Warn("无法打开HTTP日志文件:", err)
		HTTPLogger.SetOutput(os.Stdout)
	} else {
		HTTPLogger.SetOutput(io.MultiWriter(os.Stdout, httpLogFile))
	}

	logrus.Info("日志系统初始化完成")
}
