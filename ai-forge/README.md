# AI-Forge

轻量级 Go + Gin 项目骨架，用于快速搭建 AI 原生 API 平台（教学用途）。

运行（本地开发）:
1. 进入项目目录：
   cd ai-forge
2. 下载依赖并运行：
   go mod tidy
   go run ./cmd/server

访问：http://localhost:8080/ping 返回 {"message":"pong"}

使用 Docker Compose (包含 MySQL 与 Redis)：
1. 确保已安装 Docker 和 Docker Compose。
2. 在项目目录下启动服务：
   docker compose up -d
   # 这会启动 mysql (3306) 与 redis (6379)
3. （可选）复制 .env.example 到 .env 并调整：
   cp .env.example .env
4. 在不容器化 app 的情况下，默认 DB 地址为 127.0.0.1:3306，使用示例凭据（user: root / password: password）。
   启动 Go 应用：
   go run ./cmd/server

在容器内运行 Go 应用：
- 如需把 Go app 放到 compose 网络中，修改 MYSQL_HOST 为 "db"（compose 服务名），或在 container 中设置相应的环境变量：
  MYSQL_HOST=db

备注：启动后，服务会尝试连接数据库并运行 GORM AutoMigrate 来创建 users 表。
