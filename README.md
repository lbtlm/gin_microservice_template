# OSS文件管理中间件

基于 Gin 框架开发的阿里云 OSS 文件上传下载中间件项目，专为分布式客户端提供文件管理 API。

## 项目特性

- ✅ 批量文件上传到阿里云 OSS
- ✅ 分层目录结构 (app_name/user_id/file_style/)
- ✅ 多种文件类型支持 (accounts/media/avatar)
- ✅ 保持原始文件名
- ✅ 文件下载功能
- ✅ 文件删除功能
- ✅ Swagger API 文档
- ✅ 跨域支持
- ✅ 请求日志记录
- ✅ 容器化部署
- ✅ 多架构支持 (AMD64/ARM64)

## 项目结构

```
gin_middleware_oss/
├── cmd/server/      # 程序入口
│   └── main.go
├── internal/        # 内部包
│   ├── api/
│   │   ├── middleware/  # 中间件
│   │   └── v1/         # API控制器
│   ├── config/     # 配置管理
│   ├── models/     # 数据模型
│   ├── services/   # 业务逻辑
│   └── utils/      # 工具函数
├── api/swagger/    # Swagger文档
├── deployments/    # 部署配置
├── scripts/        # 构建脚本
├── go.mod         # Go模块文件
├── go.sum         # 依赖校验文件
└── README.md      # 项目说明
```

## 技术栈

- **框架**: Gin Web Framework
- **存储**: 阿里云 OSS (对象存储服务)
- **文档**: Swagger/OpenAPI
- **日志**: Logrus
- **配置**: 环境变量管理
- **部署**: Docker 容器化

## 文件存储结构

上传的文件按照以下目录结构存储：

```
OSS存储桶/
└── {app_name}/           # 应用名称（一级目录）
    └── {user_id}/        # 用户ID（二级目录）
        ├── accounts/     # accounts类型文件
        ├── avatars/      # avatar类型文件
        └── media/        # media类型文件
```

### 文件类型说明

- **accounts**: 账户相关文件，支持多种格式
- **avatar**: 头像文件，支持图片格式 .jpg, .jpeg, .png, .gif, .bmp, .webp
- **media**: 媒体文件，支持多种文件格式（图片、文档、音视频等）

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd gin_middleware_oss
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境变量

复制 `deployments/.env.example` 为 `deployments/.env.production` 并填入你的配置：

```bash
cp deployments/.env.example deployments/.env.production
```

编辑 `.env.production` 文件：

```env
# OSS配置 (必需)
OSS_ENDPOINT=oss-cn-hangzhou.aliyuncs.com
OSS_ACCESS_KEY_ID=your-access-key-id
OSS_ACCESS_KEY_SECRET=your-access-key-secret
OSS_BUCKET_NAME=your-bucket-name

# 服务器配置
SERVER_PORT=8083
SERVER_HOST=0.0.0.0
SERVER_DOMAIN=localhost:8083
CORS_ALLOWED_ORIGINS=*

# Docker配置
APP_PORT=8083
PROJECT_NAME=oss_service
```

### 4. 本地开发运行

```bash
# 生成Swagger文档
swag init -g cmd/server/main.go -o api/swagger

# 启动服务
go run cmd/server/main.go
```

### 5. 容器化部署

```bash
# 多架构镜像构建
./scripts/multi-arch-build.sh

# 本地测试部署
./deployments/smart-redis-deploy.sh

# 生产部署
./deployments/production-safe-deploy.sh
```

### 6. 访问服务

- **API 文档**: http://localhost:8083/swagger/index.html
- **健康检查**: http://localhost:8083/health
- **基础测试**: http://localhost:8083/ping

## API 接口

### 系统接口

| 方法 | 路径 | 描述 |
|-----|------|------|
| GET | `/ping` | 系统测试接口 |
| GET | `/health` | 健康检查接口 |

### 文件管理接口

| 方法 | 路径 | 描述 |
|-----|------|------|
| POST | `/api/oss/upload` | 批量文件上传 |
| GET | `/api/oss/download/{filepath}` | 文件下载 |
| DELETE | `/api/oss/delete/{filepath}` | 删除文件 |

## 使用示例

### 批量文件上传（FastAPI客户端调用）

```bash
curl -X POST \
  http://localhost:8083/api/oss/upload \
  -H 'Content-Type: multipart/form-data' \
  -F 'app_name=my_app' \
  -F 'user_id=user123' \
  -F 'file_style=media' \
  -F 'file_list=@/path/to/file1.jpg' \
  -F 'file_list=@/path/to/file2.png'
```

### Python FastAPI 客户端示例

```python
import requests
import aiofiles
from fastapi import FastAPI, File, UploadFile, Form
from typing import List

app = FastAPI()

@app.post("/upload-to-oss/")
async def upload_to_oss(
    app_name: str = Form(...),
    user_id: str = Form(...),
    file_style: str = Form(...),  # accounts/media/avatar
    files: List[UploadFile] = File(...)
):
    # 准备上传数据
    upload_data = {
        'app_name': app_name,
        'user_id': user_id,
        'file_style': file_style
    }
    
    # 准备文件数据
    files_data = []
    for file in files:
        content = await file.read()
        files_data.append(('file_list', (file.filename, content, file.content_type)))
    
    # 发送到OSS服务
    response = requests.post(
        "http://localhost:8083/api/oss/upload",
        data=upload_data,
        files=files_data
    )
    
    return response.json()
```

### 文件下载

```bash
curl -X GET \
  http://localhost:8083/api/oss/download/my_app/user123/media/example.jpg \
  -o downloaded-file.jpg
```

### 文件删除

```bash
curl -X DELETE \
  http://localhost:8083/api/oss/delete/my_app/user123/media/example.jpg
```

## 请求参数说明

### 批量上传接口参数

- **app_name** (必填): 应用名称，用于一级目录分类
- **user_id** (必填): 用户ID，用于二级目录分类
- **file_style** (必填): 文件类型，可选值：
  - `accounts`: 账户相关文件
  - `media`: 媒体文件，存储在 media/ 目录
  - `avatar`: 头像文件，存储在 avatars/ 目录
- **file_list** (必填): 文件列表，支持多文件上传

## 响应格式

### 批量上传响应示例

```json
{
  "code": 200,
  "message": "总共 2 个文件，成功 2 个，失败 0 个",
  "total_files": 2,
  "success_count": 2,
  "failed_count": 0,
  "results": [
    {
      "original_name": "example.jpg",
      "success": true,
      "message": "上传成功",
      "file_info": {
        "app_name": "my_app",
        "user_id": "user123",
        "file_style": "media",
        "original_name": "example.jpg",
        "file_name": "example.jpg",
        "file_size": 1024000,
        "file_type": ".jpg",
        "file_path": "my_app/user123/media/example.jpg",
        "download_url": "https://bucket.oss-cn-hangzhou.aliyuncs.com/my_app/user123/media/example.jpg"
      },
      "download_url": "https://bucket.oss-cn-hangzhou.aliyuncs.com/my_app/user123/media/example.jpg"
    }
  ]
}
```

## 支持的文件类型

### Accounts 类型
- 支持多种文档和数据文件格式

### Avatar 类型
- **.jpg, .jpeg**: JPEG 图片
- **.png**: PNG 图片
- **.gif**: GIF 图片
- **.bmp**: BMP 图片
- **.webp**: WebP 图片

### Media 类型
- **图片**: .jpg, .jpeg, .png, .gif, .bmp, .webp
- **文档**: .pdf, .doc, .docx, .xls, .xlsx
- **文本**: .txt, .md, .csv
- **压缩**: .zip
- **视频**: .mp4
- **音频**: .mp3, .wav

## 部署配置

### 环境变量说明

- `OSS_ENDPOINT`: OSS 地域节点
- `OSS_ACCESS_KEY_ID`: 访问密钥ID
- `OSS_ACCESS_KEY_SECRET`: 访问密钥Secret
- `OSS_BUCKET_NAME`: 存储桶名称
- `SERVER_PORT`: 服务端口，默认 8083
- `SERVER_HOST`: 服务地址，默认 0.0.0.0
- `SERVER_DOMAIN`: 外部访问域名

### Docker 部署

项目支持标准化容器部署，包含：

- 多架构镜像支持 (AMD64/ARM64)
- 智能Redis检测配置
- 生产安全部署脚本
- 健康检查机制

详细部署说明请参考 `deployments/README.md`

## 开发说明

### 目录说明

- `cmd/server/`: 应用程序入口
- `internal/api/`: HTTP API层，包含控制器和中间件
- `internal/services/`: 业务逻辑层，处理 OSS 相关操作
- `internal/models/`: 数据模型定义
- `internal/config/`: 配置管理
- `internal/utils/`: 工具函数
- `api/swagger/`: Swagger 生成的文档文件

### 添加新功能

1. 在 `internal/services/` 中添加业务逻辑
2. 在 `internal/api/v1/` 中添加 HTTP 处理函数
3. 在 `internal/api/v1/routes.go` 中注册新路由
4. 添加 Swagger 注释
5. 运行 `swag init -g cmd/server/main.go -o api/swagger` 更新文档

## 许可证

Apache License 2.0

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题请创建 Issue 或联系维护者。