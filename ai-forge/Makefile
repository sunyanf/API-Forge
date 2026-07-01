# AI-Forge Makefile

.PHONY: run dev docker-up docker-down test clean

# 本地开发（SQLite，零依赖启动）
run:
	go run main.go

# 开发模式（SQLite + 热重载）
dev:
	go run main.go

# Docker 启动完整环境（MySQL + Redis + App + Swagger）
docker-up:
	docker-compose down
	docker-compose build
	docker-compose up -d
	@echo ""
	@echo "✅ 所有服务已启动"
	@echo "   API:          http://localhost:8080"
	@echo "   Swagger UI:   http://localhost:8081"
	@echo "   Dashboard:    http://localhost:8080/dashboard"
	@echo ""

# 停止 Docker 服务
docker-down:
	docker-compose down

# 运行测试
test:
	go test ./... -v

# 清理构建产物
clean:
	rm -f data.db
	rm -f app.exe
	rm -rf logs/
