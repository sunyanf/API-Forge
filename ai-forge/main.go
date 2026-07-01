// 文件名: main.go
// 作用: AI-Forge 平台的启动入口文件
// 说明: 这是整个项目的大脑，它负责加载配置、连接数据库、
//       注册所有 API 路由，然后启动 HTTP 服务器。
//
// 运行方式: go run main.go
// 访问地址: http://localhost:8080

package main

import (
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin" // Gin Web 框架 —— 处理 HTTP 请求/响应

	"github.com/sunyanf/ai-forge/config"      // 配置模块 —— 读取环境变量
	"github.com/sunyanf/ai-forge/internal/db"  // 数据库模块 —— 连接和管理数据库
	"github.com/sunyanf/ai-forge/model"        // 数据模型 —— 定义 User、APILog 等结构体
	rootHandler "github.com/sunyanf/ai-forge/handler"   // 处理器层 —— 处理具体的 API 请求
	"github.com/sunyanf/ai-forge/middleware"   // 中间件层 —— 认证、日志、限流等
)

// ensureLogFile 确保日志文件存在，如果目录不存在则自动创建
// 参数 path: 日志文件的完整路径（如 ./logs/app.log）
// 返回值: 文件指针和可能的错误
func ensureLogFile(path string) (*os.File, error) {
	// 创建日志文件所在的目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	// 打开（或创建）日志文件，以追加模式写入
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// main 是程序的入口函数，整个应用从这里开始运行
func main() {
	// ========== 第1步：加载配置 ==========
	// 从环境变量中读取各项配置（端口、数据库地址等）
	config.Load()

	// ========== 第2步：配置日志文件 ==========
	// 所有的日志（包括 Gin 框架的日志）都写入到指定文件
	logPath := config.C.AppLog
	f, err := ensureLogFile(logPath)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}
	defer f.Close() // 程序退出时自动关闭日志文件
	// 设置标准库 log 的输出目标
	log.SetOutput(f)
	// 让 Gin 框架也把日志同时输出到文件和命令行
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(f, os.Stderr)

	// ========== 第3步：连接数据库 ==========
	// 默认使用 SQLite（零依赖，本地开发直接用）
	// Docker 环境中自动切换为 MySQL
	if err := db.Connect(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// ========== 第4步：自动创建/更新数据库表 ==========
	// GORM 会根据 model 中定义的结构体自动创建表
	// 如果表已经存在，会自动添加缺少的列
	if err := db.AutoMigrate(&model.User{}, &model.APILog{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// ========== 第5步：创建 Gin 路由引擎 ==========
	// gin.Default() 默认带了 Logger 和 Recovery 中间件
	r := gin.Default()
	// 添加自定义请求日志中间件（记录每个请求的详细信息）
	r.Use(middleware.RequestLogger())

	// ========== 第6步：注册所有 API 路由 ==========

	// --- 系统级路由 ---
	r.GET("/health", rootHandler.Health) // 健康检查：服务是否正常运行
	r.GET("/ping", rootHandler.Ping)     // 简单连通性测试

	// --- API v1 路由组 ---
	// 所有业务 API 都在 /api/v1 路径下
	api := r.Group("/api/v1")
	{
		// ---- 公开接口（无需登录） ----
		api.POST("/register", rootHandler.Register) // 用户注册
		api.POST("/login", rootHandler.Login)       // 用户登录，返回 JWT Token

		// ---- 需登录的接口（JWT 鉴权） ----
		api.GET("/me", middleware.AuthMiddleware(), rootHandler.Me) // 获取当前用户信息

		// ---- 免费服务接口（有限流保护） ----
		// RateLimitMiddleware(60, time.Minute) 表示：每分钟最多 60 次请求
		api.GET("/ip/location", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetIPLocation)      // IP 归属地查询
		api.GET("/image/random", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetRandomImage)    // 随机图片生成
		api.GET("/image/redirect", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetRandomImageRedirect) // 图片重定向

		// ---- VIP 接口（需登录 + 鉴权角色） ----
		authGroup := api.Group("")
		authGroup.Use(middleware.AuthMiddleware()) // 需要 JWT Token
		{
			authGroup.POST("/ai/content", rootHandler.GenerateAIContent) // AI 内容创作（VIP 专享）
			authGroup.GET("/usage/stats", rootHandler.GetUsageStats)     // 查看使用量统计
			authGroup.POST("/change-password", rootHandler.ChangePassword) // 修改密码

			// ---- 管理员接口 ----
			authGroup.GET("/admin/users", rootHandler.GetUserList)       // 用户列表
			authGroup.POST("/admin/upgrade", rootHandler.UpgradeUser)    // 升级/修改用户角色
		}
	}

	// --- 文档和前端路由 ---
	r.StaticFile("/docs/openapi.yaml", "./openapi.yaml")                // OpenAPI 规范文件
	r.GET("/docs", func(c *gin.Context) { c.File("./docs/swagger_index.html") }) // Swagger UI
	r.GET("/dashboard", func(c *gin.Context) { c.File("./docs/dashboard.html") }) // 前端管理界面

	// ========== 第7步：启动 HTTP 服务器 ==========
	port := config.C.AppPort
	if port == "" {
		port = "8080" // 默认端口
	}

	// 启动前检查端口是否被占用
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("❌ 端口 %s 已被占用，无法启动服务", port)
		log.Printf("")
		log.Printf("解决方法（在 PowerShell 中执行）：")
		log.Printf("  netstat -ano | findstr \":%s\" | findstr \"LISTENING\"", port)
		log.Printf("  记下最后一列的 PID 数字，然后执行：")
		log.Printf("  taskkill /PID <那个数字> /F")
		log.Printf("")
		log.Printf("或者直接双击运行 run.bat（自动处理）")
		os.Exit(1)
	}
	ln.Close() // 释放探测用的端口，交给 Gin 去监听

	log.Printf("AI-Forge 启动成功，监听端口: %s", port)
	log.Printf("  访问地址:")
	log.Printf("    API:       http://localhost:%s", port)
	log.Printf("    Dashboard: http://localhost:%s/dashboard", port)
	log.Printf("    Swagger:   http://localhost:%s/docs", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器异常退出: %v", err)
	}
}
