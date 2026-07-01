// 文件名: rate_limit.go
// 作用: 速率限制中间件
// 说明: 基于令牌桶算法实现对 API 的访问频率控制，
//       防止单个用户或 IP 过度调用接口。
//       免费用户 60次/分钟，VIP 用户 1000次/分钟。

package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
)

// RateLimiter 速率限制器（基于令牌桶算法）
// 每个独立标识（IP 或用户 ID）在一个时间窗口内拥有固定数量的令牌
type RateLimiter struct {
	visitors map[string]*visitor // 访问者映射表，key 为 IP 或用户 ID
	mu       sync.Mutex          // 互斥锁，保证并发安全（多个请求同时访问时不会出错）
	rate     int                 // 每个时间窗口允许的请求次数
	window   time.Duration       // 时间窗口大小（如 1 分钟）
}

// visitor 单个访问者的状态
type visitor struct {
	tokens    int       // 当前剩余的令牌数
	lastReset time.Time // 上次重置令牌的时间
}

// NewRateLimiter 创建一个新的速率限制器
// 参数 rate: 时间窗口内允许的请求次数（如 60）
// 参数 window: 时间窗口大小（如 1 分钟）
// 返回: 速率限制器指针（会自动启动后台清理协程）
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor), // 初始化访问者映射表
		rate:     rate,
		window:   window,
	}
	// 启动后台协程，定期清理超过窗口期的过期访问者记录
	// 这是为了防止内存泄漏——如果不清理，visitors 映射表会无限增长
	go rl.cleanup()
	return rl
}

// cleanup 定期（每分钟）清理过期的访问者记录
// 作为后台 goroutine 运行
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute) // 每分钟触发一次
	defer ticker.Stop()                   // 函数退出时停止 ticker
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			// 如果该访问者的窗口已经过期，删除记录释放内存
			if now.Sub(v.lastReset) > rl.window {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 判断指定 key 的请求是否被允许
// 参数 key: 访问者标识（IP 地址或用户 ID 字符串）
// 返回:
//   - bool: true 表示允许本次请求，false 表示被限流
//   - int: 本次请求后剩余的可用令牌数
//   - time.Time: 令牌重置时间（帮助客户端知道何时可以重试）
func (rl *RateLimiter) Allow(key string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	// 情况1：新访问者——初始化令牌并放行
	if !exists {
		rl.visitors[key] = &visitor{
			tokens:    rl.rate - 1, // 首次请求消耗 1 个令牌
			lastReset: now,
		}
		return true, rl.rate - 1, now.Add(rl.window)
	}

	// 情况2：时间窗口已过——重置令牌并放行
	if now.Sub(v.lastReset) > rl.window {
		v.tokens = rl.rate - 1 // 重置为满额减 1
		v.lastReset = now
		return true, rl.rate - 1, now.Add(rl.window)
	}

	// 情况3：令牌已用完——拒绝请求（返回 429 Too Many Requests）
	if v.tokens <= 0 {
		return false, 0, v.lastReset.Add(rl.window)
	}

	// 情况4：还有令牌——消耗 1 个并放行
	v.tokens--
	return true, v.tokens, v.lastReset.Add(rl.window)
}

// RateLimitMiddleware 速率限制中间件（Gin Handler 函数）
// 将 RateLimiter 封装为 Gin 框架的中间件形式
// 参数 rate: 时间窗口内允许的请求次数
// 参数 window: 时间窗口大小
// 返回: Gin 中间件函数
func RateLimitMiddleware(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		// 第一步：确定限流 key
		// 已登录用户 → 使用用户 ID
		// 未登录用户 → 使用客户端 IP 地址
		key := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok {
				// 将 uint 类型的用户 ID 转为字符串
				key = strconv.FormatUint(uint64(uid), 10)
			} else {
				// 类型转换失败则回退到 IP
				key = c.ClientIP()
			}
		}

		// 第二步：检查是否允许该请求
		allowed, remaining, resetTime := limiter.Allow(key)

		// 第三步：设置限流响应头（标准做法）
		// 客户端可根据这些头信息实现友好的重试策略
		c.Header("X-RateLimit-Limit", strconv.Itoa(rate))                      // 窗口内总配额
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))              // 当前剩余配额
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))  // 配额重置时间（Unix 时间戳）

		// 第四步：如果被限流，返回 429 错误
		if !allowed {
			response.Error(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort() // 中断请求链，不再执行后续处理器
			return
		}

		// 第五步：放行请求，继续执行下一个处理器
		c.Next()
	}
}
