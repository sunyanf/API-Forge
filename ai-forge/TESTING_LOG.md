# AI-Forge 测试与优化记录

## 测试日期：2026-06-30

### 问题排查与修复记录

#### 问题 1: 网络依赖下载失败
**现象**: 无法从 GitHub 下载依赖包，错误：`Connection was reset`

**原因**: 网络限制导致无法访问 GitHub

**修复方案**:
1. 设置 GOPROXY: `go env -w GOPROXY=https://goproxy.cn,direct`
2. 更新 go.mod 移除问题依赖 `github.com/chenzhuoyu/base64x`
3. 使用替代包 `github.com/cloudwego/base64x`

**状态**: ✅ 已修复

---

#### 问题 2: 数据库连接失败
**现象**: 程序启动时尝试连接 MySQL（127.0.0.1:3306）失败，报错：`connectex: No connection could be made because the target machine actively refused it`

**原因**: 本地没有运行 MySQL 数据库

**修复方案**:
修改 `internal/db/db.go`，添加 SQLite 支持：
- 检测环境变量 `DB_TYPE=sqlite`
- 使用 `github.com/glebarez/sqlite` 驱动
- SQLite 文件默认存储为 `./data.db`

**代码变更**:
```go
// 添加 SQLite 支持
dbType := os.Getenv("DB_TYPE")
if dbType == "sqlite" {
    dbFile := getenv("SQLITE_FILE", "./data.db")
    dbConn, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{Logger: newLogger})
    // ...
}
```

**状态**: ✅ 已修复

---

#### 问题 3: main.go 未注册新端点
**现象**: 新的 API 端点（IP 查询、随机图片等）无法访问

**原因**: main.go 只注册了基础端点，遗漏了新增的 handler

**修复方案**:
更新 main.go，添加所有路由：
```go
api.GET("/ip/location", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetIPLocation)
api.GET("/image/random", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetRandomImage)
api.GET("/image/redirect", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetRandomImageRedirect)
```

**状态**: ✅ 已修复

---

### 测试结果

#### 1. Health Check ✅
```bash
curl http://localhost:8080/health
```
**响应**:
```json
{
  "data": {
    "status": "ok",
    "version": "1.0.0",
    "timestamp": "2026-06-30T12:24:35Z",
    "database": "connected"
  }
}
```

---

#### 2. Ping ✅
```bash
curl http://localhost:8080/ping
```
**响应**:
```json
{"message": "pong"}
```

---

#### 3. 用户注册 ✅
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456","name":"Test User"}'
```
**响应**:
```json
{
  "data": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User",
    "role": "",
    "created_at": "2026-06-30T20:24:49.7478325+08:00"
  }
}
```

---

#### 4. 用户登录 ✅
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
```
**响应**:
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

---

#### 5. 获取当前用户信息（需要认证）✅
```bash
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <token>"
```
**响应**:
```json
{
  "data": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User",
    "role": "user",
    "api_key": "cf8205af-a598-46e2-bc3f-16cdea14ff39",
    "created_at": "2026-06-30T20:24:49.7478325+08:00"
  }
}
```

---

#### 6. IP 归属地查询 ✅
```bash
curl "http://localhost:8080/api/v1/ip/location?ip=8.8.8.8"
```
**响应**:
```json
{
  "data": {
    "ip": "8.8.8.8",
    "country": "美国",
    "country_code": "US",
    "region": "VA",
    "region_name": "弗吉尼亚州",
    "city": "Ashburn",
    "zip": "20149",
    "lat": 39.03,
    "lon": -77.5,
    "timezone": "America/New_York",
    "isp": "Google LLC",
    "org": "Google Public DNS",
    "as": "AS15169 Google LLC",
    "status": "success"
  }
}
```

---

#### 7. 随机图片 API ✅
```bash
curl "http://localhost:8080/api/v1/image/random?width=800&height=600"
```
**响应**:
```json
{
  "data": {
    "url": "https://picsum.photos/800/600?random=2149",
    "width": 800,
    "height": 600,
    "provider": "picsum.photos",
    "generated_at": "2026-06-30T12:25:31Z"
  }
}
```

---

### 运行方式

#### 本地开发（使用 SQLite）
```bash
cd ai-forge
export DB_TYPE=sqlite
export SQLITE_FILE=./data.db
go run main.go
```

#### Docker 部署（使用 MySQL）
```bash
cd ai-forge
docker-compose up -d
```

服务地址：
- API: http://localhost:8080
- Swagger UI: http://localhost:8081

### 优化总结

1. **数据库灵活性**: 支持 SQLite（开发）和 MySQL（生产）
2. **统一响应格式**: 所有 API 使用 `{"data": ...}` 格式
3. **速率限制**: 对公开 API 实施 60次/分钟限制
4. **错误处理**: 统一的错误响应格式
5. **健康检查**: 包含数据库连接状态的完整健康检查

### 已实现功能清单

- ✅ 用户注册和登录
- ✅ JWT Token 认证
- ✅ API Key 生成和管理
- ✅ IP 归属地查询
- ✅ 随机图片生成
- ✅ 速率限制中间件
- ✅ 统一响应格式
- ✅ 健康检查端点
- ✅ OpenAPI 文档
- ✅ Docker 配置

---

#### 问题 4: 端口 8080 被占用

**现象**: `go run main.go` 报错：
```
[ERROR] listen tcp :8080: bind: Only one usage of each socket address 
(protocol/network address/port) is normally permitted.
exit status 1
```

**原因**: 之前启动的 `go run` 或 `app.exe` 进程未关闭，占用了 8080 端口。

**修复方法**:
```bash
# 第一步：找到占用端口的进程
netstat -ano | findstr ":8080" | findstr "LISTENING"

# 第二步：强制结束该进程（替换 PID）
taskkill /PID <进程ID> /F

# 第三步：重新启动
go run main.go
```

**状态**: ✅ 已修复

**预防措施**: 每次重启前先执行 `taskkill` 清理旧进程。

---

#### 问题 5: 工作目录错误

**现象**: 在 Git Bash 或 PowerShell 中执行 `go run main.go` 报错：
```
GetFileAttributesEx main.go: The system cannot find the file specified.
```

**原因**: Bash 环境的工作目录未自动切换到 `ai-forge` 子目录。根目录的 `go.mod` 已被删除后，`go` 命令无法找到模块。

**修复方法**:
```bash
# 务必在 ai-forge 目录下运行
cd d:\workspace\Projects\API-Forge\ai-forge
go run main.go
```

**状态**: ✅ 已修复

---

### 下一步建议

1. 添加 API Key 刷新/撤销端点
2. 实现 API 调用日志记录并接入 Dashboard
3. 在 Dashboard 中添加管理员面板
4. 集成真实大模型 API（替换 Mock）
5. 完善单元测试覆盖率
