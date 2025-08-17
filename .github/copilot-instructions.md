# GitHub Copilot Instructions

## 🌍 语言偏好设置 (Language Preference)

**重要提醒**: 本项目所有生成的代码、注释、文档、日志消息和变量命名都应使用**中文**。

### 代码生成规范

- **注释**: 所有代码注释必须使用中文
- **错误消息**: 所有错误提示和日志消息使用中文
- **变量命名**: 优先使用中文拼音或有意义的中文词汇
- **API 文档**: Swagger 注释使用中文描述
- **配置说明**: 配置文件和环境变量说明使用中文

### 示例模式

```go
// ✅ 正确示例
func (s *OSSService) 上传文件(文件名 string) error {
    // 检查文件是否存在
    if 文件名 == "" {
        return fmt.Errorf("文件名不能为空")
    }
    LogInfo("开始上传文件: " + 文件名)
    return nil
}

// ❌ 避免纯英文
func (s *OSSService) UploadFile(filename string) error {
    // Check if file exists
    return errors.New("file not found")
}
```

### 文档生成

所有 AI 生成的文档、README 更新、代码说明都应使用中文，保持与项目现有中文文档风格一致。

---

## Project Overview

OSS 文件管理中间件 - A Gin-based microservice for Alibaba Cloud OSS file management with multi-client distribution support.

## Architecture & File Organization

### Core Structure Pattern

```
cmd/server/           # Application entry point
internal/api/         # HTTP layer (controllers, middleware)
  └── v1/            # API version 1 endpoints
  └── middleware/    # Custom middleware (CORS, logging)
internal/services/    # Business logic layer (OSS operations)
internal/config/      # Environment-based configuration
internal/models/      # Data transfer objects
internal/utils/       # Shared utilities
api/swagger/         # Auto-generated API documentation
```

### Key Design Patterns

**Three-Tier File Structure**: Files are organized as `{app_name}/{user_id}/{file_style}/filename`

- `file_style` determines subdirectory: `accounts`→`sessions/`, `avatar`→`avatars/`, `media`→`media/`
- Original filenames are preserved, no UUID renaming

**Service Layer Pattern**: Controllers delegate to services (`OSSService`), never handle business logic directly

- Controllers: HTTP request/response handling, parameter validation
- Services: OSS SDK interactions, file operations, business rules

**Configuration Management**: Environment variables loaded via `godotenv` with fallback defaults

- OSS credentials: `OSS_ACCESS_KEY_ID`, `OSS_ACCESS_KEY_SECRET`, `OSS_BUCKET_NAME`, `OSS_ENDPOINT`
- Server config: `SERVER_PORT`, `SERVER_HOST`, `SERVER_DOMAIN`

## Development Workflow

### Standard Commands

```bash
make swagger           # Regenerate API docs (do this after API changes)
make build-multi       # Multi-arch Docker build (ARM64 + AMD64)
make deploy-dev        # Smart Redis detection + local development
make deploy-prod       # Production deployment (never deletes data volumes)
```

### File Type Validation Rules

- `accounts`: Only `.json`, `.session` files
- `avatar`: Only `.png` files
- `media`: Images (`.jpg`, `.png`, `.gif`), docs (`.pdf`, `.txt`), video (`.mp4`), etc.

### API Patterns

All endpoints require structured logging with `app_name` and `user_id` for audit trails:

- Upload: `POST /api/oss/upload` - multipart form with `file_list[]`
- Download: `GET /api/oss/download/{filepath}?app_name=X&user_id=Y`
- Delete: `DELETE /api/oss/delete/{filepath}?app_name=X&user_id=Y`

## Deployment Strategy - MANDATORY

### Multi-Architecture Support

**ALWAYS** use `./scripts/multi-arch-build.sh` for container builds - supports both ARM64 and AMD64.
Never build single-architecture images.

### Production Deployment Process

1. Local multi-arch build: `make build-multi`
2. Test deployment: `make deploy-dev`
3. Production deployment: `./deployments/production-safe-deploy.sh`
4. Health verification: `curl http://localhost:8080/health`

### Critical Safety Rules

- **NEVER** delete existing data volumes in production
- **ALWAYS** wait for health checks to pass before considering deployment successful
- Use `docker-compose.smart.yml` - it includes Redis dependency management
- The deployment scripts include built-in data protection

## Code Modification Guidelines

### 🇨🇳 中文编码实践

#### 函数和方法命名

```go
// ✅ 推荐：使用中文拼音或有意义的中文标识
func 批量上传文件() {}
func 获取文件信息() {}
func 删除OSS文件() {}

// ✅ 可接受：英文但配中文注释
func BatchUpload() {
    // 批量上传文件到OSS存储
}
```

#### 错误处理和日志

```go
// ✅ 所有错误消息使用中文
return fmt.Errorf("OSS配置不完整，请检查环境变量")
middleware.LogError("上传文件失败", err)
middleware.LogInfo("用户%s开始上传文件：%s", userID, 文件名)

// ❌ 避免纯英文错误消息
return errors.New("configuration invalid")
```

#### Swagger API 文档

```go
// ✅ 使用中文描述API
// @Summary 批量上传文件到OSS
// @Description 批量上传文件到阿里云OSS，支持分布式客户端调用
// @Tags 文件管理
// @Param app_name formData string true "应用名称（一级目录）"
```

### Adding New Endpoints

1. Add handler to `internal/api/v1/`
2. Register route in `internal/api/v1/routes.go`
3. Add business logic to `internal/services/`
4. Add Swagger comments (`@Summary`, `@Description`, `@Param`, `@Success`)
5. Run `make swagger` to update documentation

### Configuration Changes

- Add environment variables to `internal/config/config.go`
- Update `deployments/docker-compose.smart.yml` environment section
- Document in README.md environment variables section

### Logging Requirements

Use structured logging via `middleware.LogInfo()`, `middleware.LogError()`:

- File operations: Include app_name, user_id, filename, file_size
- HTTP requests: Automatically logged via `LoggerMiddleware()`
- Errors: Always log with context (`client_ip`, `filepath`, etc.)

## Integration Points

### OSS SDK Usage

- Client initialization happens once in `services.NewOSSService()`
- File path generation follows the three-tier pattern
- Always check file existence before operations
- Use `bucket.PutObject()` for uploads, `bucket.GetObject()` for downloads

### Swagger Documentation

- Auto-generated from code comments in controllers
- Endpoint: `/swagger/index.html`
- Regenerated via `swag init -g cmd/server/main.go -o api/swagger`

### Health Checks

- `/health`: Basic service health
- `/ping`: Simple connectivity test
- Docker health check: Uses wget against `/health` endpoint

## Common Pitfalls to Avoid

1. **Don't** skip the multi-arch build step - deployment will fail on different architectures
2. **Don't** hardcode file paths - use `generateFilePath()` method
3. **Don't** modify deployment scripts without understanding data protection implications
4. **Don't** forget to update Swagger docs after API changes
5. **Don't** bypass the structured logging patterns - they're required for audit compliance

## Testing Strategy

- Health check endpoints for integration testing
- File upload/download round-trip tests
- Multi-client simulation via different `app_name/user_id` combinations
- Container deployment testing on both ARM64 and AMD64

This project prioritizes data safety, multi-architecture compatibility, and audit-compliant logging over rapid iteration.
