package services

import (
	"context"
	"fmt"

	"gin_middleware_oss/internal/config"

	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

// ConsulRegistry Consul 服务注册器
type ConsulRegistry struct {
	client    *api.Client
	config    *config.Config
	serviceID string
}

// NewConsulRegistry 创建新的 Consul 注册器
func NewConsulRegistry(cfg *config.Config) (*ConsulRegistry, error) {
	// 创建 Consul 客户端配置
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.Consul.Address

	if cfg.Consul.Token != "" {
		consulConfig.Token = cfg.Consul.Token
	}

	// 创建 Consul 客户端
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	registry := &ConsulRegistry{
		client:    client,
		config:    cfg,
		serviceID: generateServiceID(cfg),
	}

	return registry, nil
}

// Register 注册服务到 Consul
func (r *ConsulRegistry) Register(ctx context.Context) error {
	service := &api.AgentServiceRegistration{
		ID:      r.serviceID,
		Name:    r.config.App.Name,
		Address: r.config.Service.Address,
		Port:    r.config.Service.Port,
		Tags:    []string{"oss-file", "microservice", "file-transfer"},
		Meta: map[string]string{
			// === 网关路由核心 ===
			"route_prefixes": r.config.Consul.Meta.RoutePrefix,
			"strip_prefix":   r.config.Consul.Meta.StripPrefix,
			"base_path":      r.config.Consul.Meta.BasePath,

			// === 服务标识 ===
			"service_type": r.config.Consul.Meta.ServiceType,
			"version":      r.config.Consul.Meta.Version,
			"framework":    r.config.Consul.Meta.Framework,
			"language":     r.config.Consul.Meta.Language,

			// === 监控端点 ===
			"health_path":       r.config.Consul.Meta.HealthPath,
			"metrics_path":      r.config.Consul.Meta.MetricsPath,
			"metrics_namespace": r.config.Consul.Meta.MetricsNamespace,
			"info_path":         r.config.Consul.Meta.InfoPath,

			// === 负载均衡 ===
			"weight":     r.config.Consul.Meta.Weight,
			"lb_policy":  r.config.Consul.Meta.LbPolicy,
			"timeout_ms": r.config.Consul.Meta.TimeoutMs,
			"retries":    r.config.Consul.Meta.Retries,

			// === 依赖关系 ===
			"depend_postgres": r.config.Consul.Meta.DependPostgres,
			"depend_redis":    r.config.Consul.Meta.DependRedis,
			"depend_rabbitmq": r.config.Consul.Meta.DependRabbitmq,
			"depend_mysql":    r.config.Consul.Meta.DependMysql,

			// === 原有字段保留 ===
			"environment": r.config.App.Env,
			"scheme":      r.config.Service.Scheme,
		},
		Check: &api.AgentServiceCheck{
			HTTP:                           r.config.GetConsulHealthCheckURL(),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "60s",
		},
	}

	// 注册服务
	err := r.client.Agent().ServiceRegister(service)
	if err != nil {
		return fmt.Errorf("注册服务到 Consul 失败: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"service_id":      r.serviceID,
		"service_name":    r.config.App.Name,
		"service_address": r.config.Service.Address,
		"service_port":    r.config.Service.Port,
		"consul_address":  r.config.Consul.Address,
		"health_check":    r.config.GetHealthCheckURL(),
		"metrics_path":    "/metrics",
	}).Info("OSS文件服务已成功注册到 Consul")

	return nil
}

// Deregister 从 Consul 注销服务
func (r *ConsulRegistry) Deregister(ctx context.Context) error {
	err := r.client.Agent().ServiceDeregister(r.serviceID)
	if err != nil {
		return fmt.Errorf("从 Consul 注销服务失败: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"service_id":   r.serviceID,
		"service_name": r.config.App.Name,
	}).Info("OSS文件服务已从 Consul 注销")

	return nil
}

// GetServiceID 获取服务 ID
func (r *ConsulRegistry) GetServiceID() string {
	return r.serviceID
}

// IsHealthy 检查 Consul 连接是否健康
func (r *ConsulRegistry) IsHealthy(ctx context.Context) bool {
	// 尝试获取 Consul 状态
	_, err := r.client.Status().Leader()
	if err != nil {
		logrus.WithError(err).Warn("Consul 健康检查失败")
		return false
	}

	return true
}

// GetServiceInfo 获取服务信息
func (r *ConsulRegistry) GetServiceInfo(ctx context.Context) (*api.AgentService, error) {
	services, err := r.client.Agent().Services()
	if err != nil {
		return nil, fmt.Errorf("获取服务信息失败: %w", err)
	}

	service, exists := services[r.serviceID]
	if !exists {
		return nil, fmt.Errorf("服务 %s 未找到", r.serviceID)
	}

	return service, nil
}

// UpdateHealthCheck 更新健康检查状态
func (r *ConsulRegistry) UpdateHealthCheck(ctx context.Context, status string, output string) error {
	err := r.client.Agent().UpdateTTL("service:"+r.serviceID, output, status)
	if err != nil {
		return fmt.Errorf("更新健康检查状态失败: %w", err)
	}

	return nil
}

// generateServiceID 生成唯一的服务 ID
func generateServiceID(cfg *config.Config) string {
	// 格式: serviceName-address-port
	return fmt.Sprintf("%s-%s-%d",
		cfg.App.Name,
		cfg.Service.Address,
		cfg.Service.Port,
	)
}

// ValidateConfig 验证 Consul 配置
func ValidateConfig(cfg *config.Config) error {
	if !cfg.IsConsulEnabled() {
		return fmt.Errorf("Consul 未启用")
	}

	if cfg.Consul.Address == "" {
		return fmt.Errorf("Consul 地址不能为空")
	}

	if cfg.Service.Address == "" {
		return fmt.Errorf("服务地址不能为空")
	}

	if cfg.Service.Port <= 0 {
		return fmt.Errorf("服务端口无效: %d", cfg.Service.Port)
	}

	if cfg.Service.Scheme != "http" && cfg.Service.Scheme != "https" {
		return fmt.Errorf("服务协议必须是 http 或 https: %s", cfg.Service.Scheme)
	}

	return nil
}

// GetConsulInfo 获取 Consul 连接信息
func (r *ConsulRegistry) GetConsulInfo() map[string]interface{} {
	return map[string]interface{}{
		"address":    r.config.Consul.Address,
		"enabled":    r.config.Consul.Enabled,
		"service_id": r.serviceID,
		"service": map[string]interface{}{
			"name":    r.config.App.Name,
			"address": r.config.Service.Address,
			"port":    r.config.Service.Port,
			"scheme":  r.config.Service.Scheme,
		},
		"endpoints": map[string]string{
			"health":  r.config.GetHealthCheckURL(),
			"metrics": r.config.GetMetricsURL(),
		},
		"meta": map[string]string{
			"service_type": r.config.Consul.Meta.ServiceType,
			"version":      r.config.Consul.Meta.Version,
			"framework":    r.config.Consul.Meta.Framework,
			"language":     r.config.Consul.Meta.Language,
		},
	}
}
