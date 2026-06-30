# AI‑Forge — 技术架构、数据库 ER 图 与 RESTful API 设计

版本：初稿
技术栈：Go (Gin) 后端 | MySQL (关系型存储) | Redis (缓存/限流) | 前端：Vue 或 React（任选其一）
目标读者：初级开发者

---

## 说明
本文件是为“AI‑Forge（为独立开发者打造的 AI 原生 API 工坊）”准备的完整设计文档，包含：技术架构说明、ER 图、MySQL 建表 DDL、Redis 使用建议、以及按资源整理的 RESTful API 列表和示例。文档面向初级开发者，以易懂的方式逐步说明实现要点与调用流程。

目录
- 项目概述
- 高层技术架构（说明 + 图）
- 组件与数据流解释（逐步）
- 数据模型与 ER 图
- MySQL 建表 DDL（关键表）
- Redis 使用场景与键设计
- RESTful API 列表（资源、路径、示例请求/响应）
- 常见使用序列（示例：从创建项目到调用 endpoint）
- 实现建议

---

## 项目概述
AI‑Forge 提供给独立开发者快速打造 AI 原生 HTTP API 的平台骨架。核心目标是让开发者能定义 endpoint（路由 + 模型配置）、生成 API Key/SDK、并在运行时调用外部或内部模型，辅以计量、速率限制与日志监控。

---

## 高层技术架构（说明）
主要模块：
- Web/API 层：Go + Gin，负责路由、认证、校验、编排模型调用
- 模型适配层：封装 OpenAI/其他模型 HTTP 客户端（统一配置）
- 缓存/速率层：Redis，用于速率限制、缓存热点响应、分布式锁
- 持久层：MySQL 存储用户、项目、endpoint 配置、调用日志与计费数据
- 前端：Vue/React，提供 Dashboard、Playground、文档
- 后台任务：异步任务（结算、备份、日志归档）

文字说明：
1. 浏览器或 SDK 发起请求到 Go/Gin API。  
2. API 做认证（API Key / JWT），读取配置（MySQL）并可能命中 Redis 缓存。  
3. 若需调用模型，走模型适配层（可抽象为接口，便于切换）。  
4. 请求与耗费记录入 requests_log，用于计费与监控。  
5. 异步任务负责账单计算、邮件告警、日志归档等，避免阻塞请求路径。

---

## 组件与数据流（逐步）
1. 用户注册并创建 Project（项目）
2. 在 Project 中创建 Endpoint（定义路由、方法、输入/输出 schema、模型配置）
3. 生成 API Key（可配置速率、配额）或使用 JWT 登录后在 Playground 调试
4. 客户端调用 /invoke，API 验证 Key -> 校验输入 -> 检查速率（Redis） -> 尝试从缓存命中 -> 如未命中，则调用模型适配层 -> 返回给客户端并记录日志与 token 消耗
5. 定期任务统计使用量并生成账单

---

## 数据模型与 ER 图（简化）
实体：
- users — 拥有账号信息
- projects — 用户创建的项目
- endpoints — 项目下的 API 定义
- api_keys — 用于外部调用的秘钥
- requests_log — 每次调用记录（便于计费/调试）
- usage_metrics — 汇总每日使用数据
- teams — 协作成员
- billing_invoices — 账单记录

简化 ER（文本表示）：
- 一个 user 可以拥有多个 project。
- 每个 project 有多个 endpoints 与 api_keys。
- 每次请求产生一条 requests_log 记录并计入 usage_metrics。

---

## MySQL 建表 DDL（关键表示例）
注意：示例包含必要字段；在实现时应根据业务补充索引、约束与分区策略。

1) users
```sql
CREATE TABLE users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  name VARCHAR(128),
  role ENUM('user','admin') DEFAULT 'user',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_login TIMESTAMP NULL
);
```

2) projects
```sql
CREATE TABLE projects (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  owner_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  description TEXT,
  billing_plan VARCHAR(64) DEFAULT 'free',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);
```

3) api_keys
```sql
CREATE TABLE api_keys (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  project_id BIGINT NOT NULL,
  key_hash CHAR(64) NOT NULL,
  name VARCHAR(128),
  scopes JSON,
  daily_quota INT DEFAULT 0,
  rate_limit_per_minute INT DEFAULT 60,
  expires_at TIMESTAMP NULL,
  revoked BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX idx_api_keys_project ON api_keys(project_id);
```

4) endpoints
```sql
CREATE TABLE endpoints (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  project_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  route VARCHAR(255) NOT NULL,
  method VARCHAR(10) DEFAULT 'POST',
  input_schema JSON,
  output_schema JSON,
  model_config JSON,
  version VARCHAR(32) DEFAULT 'v1',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX idx_endpoints_project ON endpoints(project_id);
```

5) requests_log
```sql
CREATE TABLE requests_log (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  endpoint_id BIGINT,
  project_id BIGINT,
  api_key_id BIGINT,
  client_ip VARCHAR(64),
  input_payload MEDIUMTEXT,
  output_payload MEDIUMTEXT,
  tokens_in INT DEFAULT 0,
  tokens_out INT DEFAULT 0,
  duration_ms INT DEFAULT 0,
  status VARCHAR(32),
  error TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE SET NULL,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL
);
CREATE INDEX idx_requests_project ON requests_log(project_id, created_at);
```

6) usage_metrics（日汇总）
```sql
CREATE TABLE usage_metrics (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  date DATE NOT NULL,
  project_id BIGINT NOT NULL,
  requests INT DEFAULT 0,
  tokens INT DEFAULT 0,
  cost_estimate DECIMAL(10,4) DEFAULT 0.0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY (date, project_id),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

7) teams
```sql
CREATE TABLE teams (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  project_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role ENUM('owner','developer','viewer') DEFAULT 'viewer',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

8) billing_invoices
```sql
CREATE TABLE billing_invoices (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  project_id BIGINT NOT NULL,
  period_start DATE,
  period_end DATE,
  amount DECIMAL(12,2) DEFAULT 0.00,
  status ENUM('pending','paid','failed') DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

提示：对文本字段（input_payload / output_payload）可考虑分表或归档以防增长过快；requests_log 可能需要按月分区。

---

## Redis 使用场景与键设计（建议）
用途：
- 速率限制计数器（固定窗口或滑动令牌桶）
- 缓存模型响应（短期缓存相同请求）
- 分布式锁（部署/结算任务）
- 临时会话（Playground 编辑会话）

示例键：
- 速率计数器：rate:{api_key_id}:{YYYYMMDDHHmm}
- 日累计：usage:project:{project_id}:{YYYYMMDD}
- 模型缓存：cache:endpoint:{endpoint_id}:hash({input_json})
- 锁：lock:billing:{project_id}

速率限制伪流程（令牌桶简化）：
1. 每分钟向 tokens:{api_key} 中放入 N 令牌（上限 = rate_limit）。
2. 请求来时用 LUA 脚本原子地检查并消费令牌；消费失败则返回 429。

---

## RESTful API 列表（关键接口）
说明：示例均使用 JSON，鉴权用 Authorization: Bearer <JWT>（管理端）或 X-API-Key: <key>（对外调用）。错误统一返回：
```json
{
  "error": "message",
  "code": "ERR_CODE"
}
```

1) 用户与认证
- POST /api/v1/auth/register
  - 描述：注册
  - 请求：
    { "email": "a@b.com", "password": "xxxx", "name": "张三" }
  - 响应 201：
    { "id": 123, "email": "a@b.com", "name": "张三" }
- POST /api/v1/auth/login
  - 描述：登录，返回 JWT
  - 请求：
    { "email":"a@b.com", "password":"xxxx" }
  - 响应 200：
    { "token":"<jwt>", "user":{"id":123,"email":"a@b.com"} }
- GET /api/v1/me
  - 描述：获取当前用户详细信息与 API Key（需 JWT）
  - 响应 200：
    { "id": 123, "email": "a@b.com", "name": "张三", "role": "user", "api_key": "sk_live_xxx", "created_at": "2026-06-03T09:00:00Z" }

2) 项目（需 JWT）
- POST /api/v1/projects
  - 请求：
    { "name":"my-ai-project", "description":"..." }
  - 响应：新 project 对象
- GET /api/v1/projects
  - 描述：列出用户 projects
- GET /api/v1/projects/:project_id
  - 描述：项目详情、配额信息

3) API Key 管理（需 JWT）
- POST /api/v1/projects/:project_id/apikeys
  - 请求：
    { "name":"client-key-1", "daily_quota":1000, "rate_limit_per_minute":60, "scopes":["invoke"] }
  - 响应：注意 — 服务器只在创建时返回明文 key，后续仅保留哈希。
    { "id":1, "key":"sk_live_XXXXX" }
- GET /api/v1/projects/:project_id/apikeys
  - 列出（不包含明文）
- POST /api/v1/projects/:project_id/apikeys/:id/revoke
  - 撤销 key

4) Endpoint 管理（需 JWT）
- POST /api/v1/projects/:project_id/endpoints
  - 请求示例：
    {
      "name":"chat",
      "route":"/v1/chat",
      "method":"POST",
      "input_schema": { /* JSON Schema */ },
      "output_schema": {},
      "model_config": { "provider":"openai","model":"gpt-4o","temperature":0.2 }
    }
  - 响应：endpoint 对象
- GET /api/v1/projects/:project_id/endpoints
- GET /api/v1/projects/:project_id/endpoints/:endpoint_id
- PUT /api/v1/projects/:project_id/endpoints/:endpoint_id
- DELETE /api/v1/projects/:project_id/endpoints/:endpoint_id

5) Endpoint 调用（对外：使用 X-API-Key）
- POST /api/v1/projects/:project_id/endpoints/:endpoint_id/invoke
  - 鉴权：X-API-Key: <key>
  - 请求举例：
    { "input": { "messages":[{"role":"user","content":"Hello"}] } }
  - 响应成功示例：
    {
      "request_id": "req_0001",
      "status": "success",
      "output": { "text": "Hi!" },
      "tokens_in": 10,
      "tokens_out": 12
    }
  - 可能返回状态：
    - 200 成功
    - 400 参数不合法（schema 验证失败）
    - 401 未授权
    - 403 Key 超出配额或被撤销
    - 429 速率限额
    - 500 模型调用失败

注意：在返回中带上速率限制 headers（便于客户端处理）：
- X-RateLimit-Limit: 60
- X-RateLimit-Remaining: 45
- X-RateLimit-Reset: 169xxxxxx (unix timestamp)

6) Playground（需 JWT）
- POST /api/v1/projects/:project_id/playground/run
  - 用于在 UI 调试 Endpoint（使用项目默认 Key 或 JWT）

7) 监控与指标（需 JWT / 管理权限）
- GET /api/v1/projects/:project_id/metrics?start=2026-05-01&end=2026-05-31
  - 返回 requests、tokens、top endpoints、错误率等汇总

8) 请求日志查询（需权限）
- GET /api/v1/projects/:project_id/requests?endpoint_id=&limit=50&since=
  - 返回 requests_log 条目（用于调试）

9) 管理后台（Admin）
- GET /api/v1/admin/usage/top-projects
- POST /api/v1/admin/projects/:id/suspend

---

## 常见使用序列（示例）
流程：创建项目 -> 创建 endpoint -> 生成 API Key -> 调用
1. 登录 -> POST /auth/login -> 得到 JWT  
2. POST /projects -> 得到 project_id  
3. POST /projects/:id/endpoints -> 定义路径 /v1/chat  
4. POST /projects/:id/apikeys -> 得到明文 key sk_xyz  
5. 客户端用 X-API-Key: sk_xyz POST /projects/:id/endpoints/:eid/invoke

在实现中要注意：
- 在创建 api_key 时只在响应中返回明文一次；数据库只保存哈希。
- 对 invoke 请求使用快速路径（认证 -> 限流 -> 校验 -> 缓存 -> 模型调用 -> 日志入库）。
- 模型调用需要超时与重试策略（例如 10s 超时，最多重试 1 次）。

---

## 实现建议（面向初级开发者）
- 分层设计：router -> handler -> service -> repository（将数据库操作封装在 repository 层）。
- 使用接口/抽象包装模型提供者（便于替换）。
- 安全实践：
  - 永远对 API Key 存哈希（例如 HMAC 或 bcrypt/sha256 + salt）。
  - 所有外部请求使用 HTTPS。
  - 敏感配置（模型 API keys）放在安全配置或 secret 管理（环境变量或 Vault）。
- 测试：为核心流程写单元测试与集成测试（例如调用 endpoint 的 happy path、速率限制场景、错误处理）。
- 日志与追踪：在请求上下文中生成 request_id，便于追踪跨系统调用。

---

## 示例 Gin 路由（简化）
```go
r := gin.Default()
api := r.Group("/api/v1")
{
  auth := api.Group("/auth")
  auth.POST("/register", RegisterHandler)
  auth.POST("/login", LoginHandler)

  proj := api.Group("/projects")
  proj.Use(AuthMiddleware()) // JWT 源
  proj.POST("", CreateProject)
  proj.GET("", ListProjects)
  proj.POST("/:project_id/endpoints", CreateEndpoint)
  proj.POST("/:project_id/apikeys", CreateAPIKey)
  proj.POST("/:project_id/endpoints/:endpoint_id/invoke", InvokeEndpoint) // 也可对外公开，使用 X-API-Key 验证
}
```

---

## 结语与下一步
文档已导出到 docs/AI-Forge-Architecture-API-Design.md。下一步可以选择：
- 生成 OpenAPI spec（可直接用于 Swagger UI）
- 在仓库中添加初始 Go Gin 项目骨架（含认证、示例 endpoint 和速率限制中间件）
- 生成 SDK 示例（Go/JS）

请告诉要继续的选项："生成 OpenAPI spec"、"生成 Go Gin 项目骨架" 或 "生成 SDK 示例"。
