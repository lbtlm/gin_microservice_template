package config

import (
	"net"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config 应用配置结构
type Config struct {
	// 应用配置
	App AppConfig

	// 日志配置
	Log LogConfig

	// 服务器配置
	Server ServerConfig

	// Consul 配置
	Consul ConsulConfig

	// 服务配置
	Service ServiceConfig
}

// AppConfig 应用基础配置
type AppConfig struct {
	Env  string
	Port string
	Name string
}

// LogConfig 日志相关配置
type LogConfig struct {
	Level  string
	Format string
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	Port               string `json:"port"`
	Host               string `json:"host"`
	Domain             string `json:"domain"`          // 用于外部访问的域名或IP
	AllowedOrigins     string `json:"allowed_origins"` // CORS允许的源，多个用逗号分隔
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	HealthCheckTimeout time.Duration
}

// ConsulConfig Consul 相关配置
type ConsulConfig struct {
	Address string
	Token   string
	Enabled bool
	Meta    ConsulMetaConfig
}

// ConsulMetaConfig Consul Meta 字段配置
type ConsulMetaConfig struct {
	// 网关路由核心
	RoutePrefix string
	StripPrefix string
	BasePath    string

	// 服务标识
	ServiceType string
	Version     string
	Framework   string
	Language    string

	// 监控端点
	HealthPath       string
	MetricsPath      string
	MetricsNamespace string
	InfoPath         string

	// 负载均衡
	Weight    string
	LbPolicy  string
	TimeoutMs string
	Retries   string

	// 依赖关系
	DependPostgres string
	DependRedis    string
	DependRabbitmq string
	DependMysql    string
}

// ServiceConfig 服务注册相关配置
type ServiceConfig struct {
	Address            string // 外部可访问地址
	Port               int
	Scheme             string
	HealthCheckAddress string // Consul 健康检查专用地址（容器内部地址）
}

var GlobalConfig *Config

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 根据环境加载对应的 .env 文件
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	var envFile string
	switch env {
	case "production":
		envFile = ".env.production"
	case "development":
		envFile = ".env.dev"
	default:
		envFile = ".env.dev"
	}

	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(envFile); err != nil {
		logrus.Warnf("未能加载环境变量文件 %s: %v", envFile, err)
		// 尝试加载默认的 .env 文件
		if err := godotenv.Load(); err != nil {
			logrus.Warnf("未能加载默认 .env 文件: %v", err)
		}
	}

	config := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
			Name: getEnv("APP_NAME", "auth-service"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Server: ServerConfig{
			Port:               getEnv("SERVER_PORT", "8080"),
			Host:               getEnv("SERVER_HOST", "0.0.0.0"),
			Domain:             getEnv("SERVER_DOMAIN", "localhost:8080"),
			AllowedOrigins:     getEnv("CORS_ALLOWED_ORIGINS", "*"),
			ReadTimeout:        parseDuration("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout:       parseDuration("SERVER_WRITE_TIMEOUT", "30s"),
			HealthCheckTimeout: parseDuration("HEALTH_CHECK_TIMEOUT", "10s"),
		},
		Consul: ConsulConfig{
			Address: getEnv("CONSUL_HTTP_ADDR", "http://127.0.0.1:8500"),
			Token:   getEnv("CONSUL_HTTP_TOKEN", ""),
			Enabled: getBool("CONSUL_ENABLED", true),
			Meta: ConsulMetaConfig{
				// 网关路由核心
				RoutePrefix: getEnv("CONSUL_ROUTE_PREFIX", "/auth-service"),
				StripPrefix: getEnv("CONSUL_STRIP_PREFIX", "true"),
				BasePath:    getEnv("CONSUL_BASE_PATH", "/"),

				// 服务标识
				ServiceType: getEnv("CONSUL_SERVICE_TYPE", "auth-service"),
				Version:     getEnv("CONSUL_VERSION", "v1.0.0"),
				Framework:   getEnv("CONSUL_FRAMEWORK", "gin"),
				Language:    getEnv("CONSUL_LANGUAGE", "go"),

				// 监控端点
				HealthPath:       getEnv("CONSUL_HEALTH_PATH", "/health"),
				MetricsPath:      getEnv("CONSUL_METRICS_PATH", "/metrics"),
				MetricsNamespace: getEnv("CONSUL_METRICS_NAMESPACE", "auth_service"),
				InfoPath:         getEnv("CONSUL_INFO_PATH", "/api/v1/stats"),

				// 负载均衡
				Weight:    getEnv("CONSUL_WEIGHT", "100"),
				LbPolicy:  getEnv("CONSUL_LB_POLICY", "weighted_round_robin"),
				TimeoutMs: getEnv("CONSUL_TIMEOUT_MS", "30000"),
				Retries:   getEnv("CONSUL_RETRIES", "2"),

				// 依赖关系
				DependPostgres: getEnv("CONSUL_DEPEND_POSTGRES", "false"),
				DependRedis:    getEnv("CONSUL_DEPEND_REDIS", "false"),
				DependRabbitmq: getEnv("CONSUL_DEPEND_RABBITMQ", "false"),
				DependMysql:    getEnv("CONSUL_DEPEND_MYSQL", "false"),
			},
		},
		Service: ServiceConfig{
			Address:            getServiceAddress(),
			Port:               getServicePort(getEnv("APP_PORT", "8080")),
			Scheme:             getEnv("SERVICE_SCHEME", "http"),
			HealthCheckAddress: getEnv("CONSUL_HEALTH_CHECK_ADDRESS", ""),
		},
	}

	GlobalConfig = config

	// 验证关键配置
	if config.App.Name == "" {
		logrus.Warn("App name is not configured, using default.")
	}
	return config, nil
}

// getEnv 获取环境变量，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// parseDuration 解析时间间隔环境变量
func parseDuration(key, defaultValue string) time.Duration {
	valueStr := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		defaultDuration, _ := time.ParseDuration(defaultValue)
		logrus.Warnf("无法解析环境变量 %s 的时间间隔: %v, 使用默认值: %s", key, err, defaultValue)
		return defaultDuration
	}
	return duration
}

// getBool 获取布尔类型环境变量
func getBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	switch valueStr {
	case "true", "True", "TRUE", "1", "yes", "Yes", "YES":
		return true
	case "false", "False", "FALSE", "0", "no", "No", "NO":
		return false
	default:
		logrus.Warnf("无法解析环境变量 %s 的布尔值: %s, 使用默认值: %t", key, valueStr, defaultValue)
		return defaultValue
	}
}

// getServiceAddress 获取服务地址，优先级：SERVICE_ADDRESS > hostname > localhost
func getServiceAddress() string {
	if addr := os.Getenv("SERVICE_ADDRESS"); addr != "" {
		return addr
	}

	// 尝试获取主机名
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		// 验证主机名不是 localhost 类似的值
		if hostname != "localhost" && hostname != "127.0.0.1" {
			return hostname
		}
	}

	// 尝试获取本机对外 IP 地址
	if ip := getOutboundIP(); ip != "" {
		return ip
	}

	return "localhost"
}

// getServicePort 获取服务端口，优先级：SERVICE_PORT > APP_PORT
func getServicePort(appPort string) int {
	if portStr := os.Getenv("SERVICE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port
		}
		logrus.Warnf("无法解析 SERVICE_PORT: %s, 使用 APP_PORT", portStr)
	}

	if port, err := strconv.Atoi(appPort); err == nil {
		return port
	}

	logrus.Warn("无法解析端口，使用默认端口 8080")
	return 8080
}

// getOutboundIP 获取本机对外 IP 地址
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// IsProduction 判断是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsDevelopment 判断是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// GetServerAddr 获取服务器监听地址
func (c *Config) GetServerAddr() string {
	return ":" + c.App.Port
}

// GetServiceURL 获取服务的完整 URL
func (c *Config) GetServiceURL() string {
	return c.Service.Scheme + "://" + c.Service.Address + ":" + strconv.Itoa(c.Service.Port)
}

// GetHealthCheckURL 获取健康检查 URL
func (c *Config) GetHealthCheckURL() string {
	return c.GetServiceURL() + "/health"
}

// GetConsulHealthCheckURL 获取 Consul 专用健康检查 URL（使用容器内部地址）
func (c *Config) GetConsulHealthCheckURL() string {
	// 如果配置了专用的健康检查地址，使用该地址
	if c.Service.HealthCheckAddress != "" {
		return c.Service.Scheme + "://" + c.Service.HealthCheckAddress + ":" + strconv.Itoa(c.Service.Port) + "/health"
	}
	// 否则使用默认地址
	return c.GetHealthCheckURL()
}

// GetMetricsURL 获取 metrics URL
func (c *Config) GetMetricsURL() string {
	return c.GetServiceURL() + "/metrics"
}

// IsConsulEnabled 判断是否启用 Consul
func (c *Config) IsConsulEnabled() bool {
	return c.Consul.Enabled
}
