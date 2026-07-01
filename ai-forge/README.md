# AI-Forge

轻量级 Go + Gin API 平台，为独立开发者打造的 AI 原生 API 工坊。

## 快速开始

### 方式 1：本地零依赖启动（推荐开发）

```bash
cd ai-forge
go run main.go
```

默认使用 SQLite，**无需安装任何数据库**，一键启动即可使用。

访问：
- API: http://localhost:8080
- Dashboard: http://localhost:8080/dashboard
- Swagger UI: http://localhost:8080/docs

### 方式 2：Docker Compose 全栈部署

```bash
cd ai-forge
docker-compose up -d
```

启动 MySQL + Redis + App + Swagger UI，完整生产环境。

访问：
- API: http://localhost:8080
- Swagger UI: http://localhost:8081
- Dashboard: http://localhost:8080/dashboard

## API 端点

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /health | 健康检查（含数据库状态） | 无 |
| GET | /ping | Ping 测试 | 无 |
| POST | /api/v1/register | 用户注册 | 无 |
| POST | /api/v1/login | 用户登录，返回 JWT | 无 |
| GET | /api/v1/me | 获取当前用户信息 | JWT |
| GET | /api/v1/ip/location | IP 归属地查询 | 无(限流) |
| GET | /api/v1/image/random | 随机图片生成 | 无(限流) |
| GET | /api/v1/image/redirect | 重定向到随机图片 | 无(限流) |

### 测试示例

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456","name":"Test"}'

# 登录
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'

# IP 查询
curl "http://localhost:8080/api/v1/ip/location?ip=8.8.8.8"

# 随机图片
curl "http://localhost:8080/api/v1/image/random?width=800&height=600"
```

## 项目结构

```
ai-forge/
├── main.go              # 应用入口
├── config/              # 配置
├── dao/                 # 数据访问层
├── handler/             # 请求处理器
├── internal/
│   ├── db/              # 数据库连接（支持 SQLite / MySQL）
│   └── router/          # 路由注册
├── middleware/           # 中间件（认证、日志、限流）
├── model/               # 数据模型
├── response/            # 统一响应格式
├── service/             # 业务逻辑层
├── docs/                # 文档和前端
│   ├── dashboard.html   # API 管理界面
│   └── swagger_index.html
├── openapi.yaml         # OpenAPI 规范
├── Dockerfile           # Docker 镜像构建
├── docker-compose.yml   # Docker 全栈编排
├── Makefile             # 常用命令
└── TESTING_LOG.md       # 测试与修复记录
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| DB_TYPE | `sqlite` | 数据库类型：`sqlite`（本地）或 `mysql`（Docker） |
| SQLITE_FILE | `./data.db` | SQLite 数据库文件路径 |
| MYSQL_HOST | `127.0.0.1` | MySQL 主机地址 |
| MYSQL_PORT | `3306` | MySQL 端口 |
| MYSQL_USER | `root` | MySQL 用户名 |
| MYSQL_PASSWORD | `password` | MySQL 密码 |
| MYSQL_DATABASE | `ai_forge` | 数据库名称 |
| PORT | `8080` | 服务端口 |
| JWT_SECRET | `dev-secret` | JWT 签名密钥 |

## 特性

- ✅ 零依赖本地启动（SQLite）
- ✅ Docker Compose 全栈部署（MySQL + Redis）
- ✅ JWT Token 认证
- ✅ API Key 鉴权
- ✅ 速率限制
- ✅ IP 归属地查询
- ✅ 随机图片生成
- ✅ 统一响应格式
- ✅ 健康检查端点
- ✅ OpenAPI 文档
- ✅ 前端管理 Dashboard
