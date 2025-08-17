# Makefile for Auth Service

# 项目信息
PROJECT_NAME := auth-service
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建配置
BINARY_NAME := auth-service
MAIN_PATH := ./cmd/server/main.go
BUILD_DIR := ./build
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := $(VERSION)

# Docker Compose 配置
DEV_COMPOSE := docker-compose.dev.yml
PROD_COMPOSE := docker-compose.pro.yml
DEV_PROJECT := auth-dev
PROD_PROJECT := auth-prod

.PHONY: help
help: ## 显示帮助信息
	@echo "Auth Service - Makefile"
	@echo "======================="
	@echo ""
	@echo "开发环境命令:"
	@echo "  dev-build            构建镜像并启动开发环境"
	@echo "  dev-stop             停用本地所有项目相关容器"
	@echo ""
	@echo "生产环境命令:"
	@echo "  pro-build            拉取镜像并部署生产环境"
	@echo "  pro-stop             生产环境停止容器（保留数据）"
	@echo ""
	@echo "镜像管理命令:"
	@echo "  build-push           构建 Linux AMD64 镜像并推送到仓库"
	@echo ""

.PHONY: dev-build
dev-build: ## 本地构建和启动开发环境
	@echo "开发环境快速构建..."
	@echo "步骤1: 整理依赖"
	@go mod tidy
	@echo "步骤2: 格式化代码"
	@go fmt ./...
	@echo "步骤3: 生成API文档"
	@$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o api/swagger
	@echo "步骤4: 构建Docker镜像"
	@source .env.dev && docker build -t $$DOCKER_IMAGE:$$DOCKER_TAG .
	@echo "步骤5: 启动开发环境"
	@docker compose -f $(DEV_COMPOSE) -p $(DEV_PROJECT) up -d
	@echo "开发环境构建和部署完成！"
	@source .env.dev 2>/dev/null || true; \
	APP_PORT=$${APP_PORT:-8086}; \
	echo "服务地址: http://localhost:$$APP_PORT"; \
	echo "Swagger文档: http://localhost:$$APP_PORT/swagger/index.html"; \
	echo "Consul管理界面: http://localhost:8500"

.PHONY: dev-stop
dev-stop: ## 停用本地所有项目相关容器
	@echo "停用本地所有容器..."
	@docker compose -f $(DEV_COMPOSE) -p $(DEV_PROJECT) down --remove-orphans -v 2>/dev/null || true
	@docker compose -f $(PROD_COMPOSE) -p $(PROD_PROJECT) down --remove-orphans -v 2>/dev/null || true
	@docker container prune -f 2>/dev/null || true
	@docker image prune -f 2>/dev/null || true
	@docker network prune -f 2>/dev/null || true
	@docker volume prune -f 2>/dev/null || true
	@echo "所有容器已停止并清理完成"

.PHONY: pro-build
pro-build: ## 生产环境拉取镜像并部署
	@echo "生产环境拉取镜像并部署..."
	@echo "步骤1: 检查配置文件"
	@if [ ! -f .env.production ]; then \
		echo ".env.production 文件不存在"; \
		echo "请先创建生产环境配置文件"; \
		exit 1; \
	fi
	@echo ".env.production 存在"
	@echo "步骤2: 拉取生产镜像"
	@set -e; \
	DOCKER_IMAGE=$$(grep DOCKER_IMAGE .env.production | cut -d= -f2) || { echo "DOCKER_IMAGE 未设置"; exit 1; }; \
	DOCKER_TAG=$$(grep DOCKER_TAG .env.production | cut -d= -f2) || { echo "DOCKER_TAG 未设置"; exit 1; }; \
	echo "拉取镜像: $$DOCKER_IMAGE:$$DOCKER_TAG"; \
	docker pull $$DOCKER_IMAGE:$$DOCKER_TAG
	@echo "步骤3: 创建必要目录"
	@mkdir -p ./data ./logs
	@echo "步骤4: 部署生产环境"
	@docker compose -f $(PROD_COMPOSE) -p $(PROD_PROJECT) up -d
	@echo "步骤5: 等待服务启动"
	@sleep 15
	@echo "步骤6: 健康检查"
	@for i in 1 2 3 4 5; do \
		echo "第$$i次健康检查..."; \
		if curl -s -f http://localhost:8086/health >/dev/null 2>&1; then \
			echo "✅ 服务启动成功！"; \
			echo "🌐 服务地址: http://8.216.34.86:8086"; \
			echo "📚 Swagger文档: http://8.216.34.86:8086/swagger/index.html"; \
			echo "🔗 Consul中央控制器: http://104.234.155.170:8500"; \
			break; \
		else \
			echo "服务尚未就绪，等待10秒..."; \
			sleep 10; \
		fi; \
		if [ $$i -eq 5 ]; then \
			echo "❌ 服务启动失败，请检查日志"; \
			docker compose -f $(PROD_COMPOSE) -p $(PROD_PROJECT) logs --tail=50; \
		fi; \
	done
	@echo "生产环境部署完成！"

.PHONY: pro-stop
pro-stop: ## 生产环境停止容器（保留数据卷）
	@echo "生产环境停止容器..."
	@echo "注意：这将停止生产环境容器但保留数据卷"
	@read -p "确认停止生产环境容器? (输入 'yes' 确认): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "停止生产环境容器..."; \
		docker compose -f $(PROD_COMPOSE) -p $(PROD_PROJECT) down --remove-orphans; \
		echo "✅ 生产环境容器已停止（数据卷已保留）"; \
	else \
		echo "取消操作"; \
	fi

.PHONY: build-push
build-push: ## 构建 Linux AMD64 镜像并推送到仓库
	@echo "构建 Linux AMD64 镜像并推送到仓库..."
	@go mod tidy
	@go fmt ./...
	@$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o api/swagger
	@source .env.dev 2>/dev/null || source .env.template; \
	echo "构建目标架构: linux/amd64"; \
	echo "镜像标签: $$DOCKER_IMAGE:$$DOCKER_TAG"; \
	docker buildx create --name multiarch-builder --use --bootstrap 2>/dev/null || true; \
	docker buildx build \
		--platform linux/amd64 \
		--tag $$DOCKER_IMAGE:$$DOCKER_TAG \
		--push \
		--file Dockerfile \
		--progress=plain \
		.
	@docker buildx rm multiarch-builder 2>/dev/null || true
	@echo "Linux AMD64 镜像构建并推送完成！"

# 默认目标
.DEFAULT_GOAL := help

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker run --rm \
		-p 8080:8080 \
		--name $(PROJECT_NAME) \
		$(PROJECT_NAME):latest

.PHONY: docker-stop
docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	@docker stop $(PROJECT_NAME) || true

.PHONY: env
env: ## 创建环境变量模板
	@echo "创建环境变量模板..."
	@if [ ! -f .env.dev ]; then \
		cp .env.template .env.dev; \
		echo "已创建 .env.dev 文件，请根据需要修改配置"; \
	else \
		echo ".env.dev 文件已存在"; \
	fi

.PHONY: install
install: ## 安装开发工具
	@echo "安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "开发工具安装完成"

# 项目初始化
.PHONY: init
init: mod env ## 初始化项目
	@echo "项目初始化完成"
	@echo "运行 'make dev' 启动开发环境"
	@echo "运行 'make help' 查看所有可用命令"
