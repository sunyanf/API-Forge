// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现 AI 内容创作（AI Content）相关的请求处理逻辑。
// 这是一个模拟（Mock）实现，不需要真实调用大模型 API，
// 通过预置模板根据用户输入的关键词生成内容，用于演示和测试。
package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
)

// AIContentRequest 定义 AI 内容生成请求的结构
type AIContentRequest struct {
	// Type 内容类型，必填。支持的值：xiaohongshu（小红书文案）、product（产品介绍）、
	//       slogan（口号/标语）、email（邮件模板）
	Type string `json:"type" binding:"required"`
	// Keyword 内容关键词，必填。模板中的 {keyword} 占位符会被替换为该值
	Keyword string `json:"keyword" binding:"required"`
}

// AIContentResponse 定义 AI 内容生成响应的结构
type AIContentResponse struct {
	// Content 生成的文本内容
	Content string `json:"content"`
	// Type 内容类型（回显请求中的类型）
	Type string `json:"type"`
	// Keyword 使用的关键词（回显请求中的关键词）
	Keyword string `json:"keyword"`
	// TokensUsed 本次生成消耗的 Token 数量（模拟值，按字符数估算）
	TokensUsed int `json:"tokens_used"`
	// Model 使用的 AI 模型名称（模拟环境为固定值）
	Model string `json:"model"`
	// CreatedAt 内容生成时间，RFC3339 格式
	CreatedAt string `json:"created_at"`
}

// GenerateAIContent 处理 AI 内容生成请求（模拟实现）
// 该接口仅限 VIP 和管理员用户使用，普通用户调用会返回 402 错误。
// 根据请求中的 type 和 keyword，从预置模板中选择并替换生成内容。
//
// 请求示例：
//
//	POST /api/ai/generate
//	Content-Type: application/json
//	Authorization: Bearer <JWT Token>
//	{"type": "xiaohongshu", "keyword": "蓝牙耳机"}
//
// 响应（成功，200 OK）：
//
//	{"content": "✨ 发现了一个超棒的蓝牙耳机！...", "type": "xiaohongshu", ...}
//
// 响应（无权限，402 Payment Required）：
//
//	{"message": "this feature requires a VIP subscription. ..."}
func GenerateAIContent(c *gin.Context) {
	// 步骤1：检查用户角色权限
	// 从 JWT 中间件注入的上下文中获取 user_role
	// 只有 vip 和 admin 角色可以调用此接口
	role, _ := c.Get("user_role")
	roleStr, _ := role.(string) // 类型断言，将 interface{} 转为 string
	if roleStr != "vip" && roleStr != "admin" {
		// 普通用户无权限，返回 402 Payment Required
		response.Error(c, 402, "this feature requires a VIP subscription. Please upgrade your account.")
		return
	}

	// 步骤2：绑定并校验请求体
	var req AIContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 缺少必填字段或 JSON 格式错误
		response.BadRequest(c, "type and keyword are required")
		return
	}

	// 步骤3：检查关键词是否有效（去除空格后不能为空）
	keyword := strings.TrimSpace(req.Keyword)
	if len(keyword) == 0 {
		response.BadRequest(c, "keyword is required")
		return
	}

	// 步骤4：使用关键词和类型生成内容
	content := generateContent(req.Type, keyword)

	// 步骤5：估算 Token 消耗量（模拟）
	// 简单按 Unicode 字符数估算，实际 AI 模型的 Token 计算更为复杂
	tokens := len([]rune(content))

	// 步骤6：构造并返回响应
	resp := AIContentResponse{
		Content:    content,                                // 生成的文本
		Type:       req.Type,                               // 回显类型
		Keyword:    keyword,                                // 回显关键词
		TokensUsed: tokens,                                 // Token 消耗量
		Model:      "ai-forge-mock-v1",                     // 模拟模型名称
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),   // 生成时间
	}

	response.OK(c, resp)
}

// generateContent 根据内容类型和关键词从模板生成文本（模拟 AI 生成）
//
// 参数：
//   - contentType: 内容类型（xiaohongshu、product、slogan、email）
//   - keyword: 用于替换模板中 {keyword} 占位符的关键词
//
// 返回值：
//   - string: 替换后的文本内容
//
// 实现细节：
// 每种类型有多个模板变体，根据关键词长度取模选择不同模板，确保输出的多样性。
// 模板中的 {keyword} 会被替换为实际关键词。
func generateContent(contentType, keyword string) string {
	// 定义各类型的模板库
	// 每种类型包含多个模板变体，增加生成内容的多样性
	templates := map[string][]string{
		// xiaohongshu：小红书风格文案，带 Emoji 表情和话题标签
		"xiaohongshu": {
			"✨ 发现了一个超棒的{keyword}！真的太值得尝试了～\n\n🔥 最近被种草了{keyword}，用了之后简直绝绝子！\n\n📝 给大家分享一下我的使用体验：\n1. 颜值超高，拿在手上很有质感\n2. 效果惊艳，完全超出预期\n3. 性价比绝了，强烈安利！\n\n#好物分享 #{keyword} #种草 #生活好物",
			"姐妹们！{keyword}真的绝了！！！😍\n\n之前一直犹豫要不要入手，现在用了之后超级后悔——后悔没早买！\n\n给大家总结几个亮点：\n🌟 颜值在线，出门回头率超高\n🌟 功能强大，用起来很顺手\n🌟 价格亲民，学生党也能冲\n\n赶紧收藏起来慢慢看吧～\n\n#{keyword} #好物推荐 #值得买",
		},
		// product：产品介绍/评测风格，偏正式商业文案
		"product": {
			"【{keyword}】核心卖点\n\n1. 差异化优势：采用创新技术，相比竞品提升40%效率\n2. 用户体验：简洁直观的操作界面，3分钟快速上手\n3. 安全可靠：多重加密保护，7×24小时稳定运行\n4. 持续迭代：每月更新优化，紧跟行业趋势\n\n目标用户：追求品质的中高端消费者",
			"💡 {keyword} — 重新定义行业标准\n\n🎯 解决痛点：\n→ 告别繁琐的操作流程\n→ 降低80%的时间成本\n→ 提升团队协作效率\n\n🏆 产品亮点：\n• 智能算法驱动，精准匹配需求\n• 模块化设计，灵活扩展\n• 云端部署，随时随地访问\n\n即刻体验，让工作更简单！",
		},
		// slogan：口号/标语生成，简洁有力
		"slogan": {
			fmt.Sprintf("「{keyword}，让生活更精彩」\n「{keyword}，不止于此」\n「选择{keyword}，选择未来」\n「{keyword} — 你的品质之选」\n「每一天，从{keyword}开始」"),
			fmt.Sprintf("• {keyword}，超越你的期待\n• {keyword}，创造无限可能\n• {keyword}，定义新高度\n• {keyword}，为卓越而生\n• {keyword}，引领潮流"),
		},
		// email：商业邮件模板
		"email": {
			fmt.Sprintf(`主题：{keyword} - 专属于您的特别推荐

尊敬的客户：

您好！

我们很高兴向您推荐{keyword}相关的最新产品和服务。

【核心优势】
• 品质保证：经过严格测试认证
• 专业服务：7×24小时贴心支持
• 限时优惠：新用户首单立享8折

【立即了解】
点击下方链接查看详情，开启您的专属体验！

如您有任何疑问，欢迎随时回复本邮件，我们将为您提供一对一服务。

祝您工作愉快！

AI-Forge 团队敬上
---
本邮件由 AI-Forge 自动生成，如有打扰敬请谅解。`),
		},
	}

	// 步骤1：根据类型查找对应的模板列表
	// 如果客户端传了不支持的类型，默认回退到 xiaohongshu 模板
	choices, ok := templates[contentType]
	if !ok {
		choices = templates["xiaohongshu"] // 兜底策略：未知类型使用小红书模板
	}

	// 步骤2：基于关键词长度选择模板（简单哈希策略）
	// 用关键词长度对模板数量取模，确保相同类型的请求能轮换到不同模板
	idx := len(keyword) % len(choices)
	template := choices[idx]

	// 步骤3：将模板中的 {keyword} 替换为实际关键词
	// strings.ReplaceAll 会替换所有匹配的占位符
	result := strings.ReplaceAll(template, "{keyword}", keyword)

	return result
}

// GetUsageStats 获取当前用户的 API 使用量统计数据（模拟）
// 返回用户的请求次数、Token 消耗、免费/VIP 配额等统计信息。
// 如果用户是 VIP 或 admin，配额会更高。
//
// 请求示例：
//
//	GET /api/usage/stats
//	Authorization: Bearer <JWT Token>
//
// 响应（普通用户，200 OK）：
//
//	{"subscription": "free", "free_remaining": 1000, "rate_limit": 60, ...}
//
// 响应（VIP 用户，200 OK）：
//
//	{"subscription": "vip", "vip_remaining": 10000, "rate_limit": 1000, ...}
func GetUsageStats(c *gin.Context) {
	// 步骤1：从上下文获取当前用户 ID
	uid, _ := c.Get("user_id")

	// 步骤2：构建默认统计数据（对应免费用户）
	resp := gin.H{
		"user_id":        uid,     // 用户 ID
		"total_requests": 0,       // 总请求次数（模拟值）
		"tokens_used":    0,       // 已使用 Token 数（模拟值）
		"free_remaining": 1000,    // 免费套餐剩余调用次数
		"vip_remaining":   10000,   // VIP 套餐剩余调用次数（默认值，对免费用户无意义）
		"rate_limit":     60,      // 每分钟请求限制（免费用户 60 次/分钟）
		"subscription":   "free",  // 订阅类型，默认为免费
	}

	// 步骤3：根据用户角色调整配额
	// VIP 和管理员拥有更高的调用限额和速率限制
	role, _ := c.Get("user_role")
	if roleStr, ok := role.(string); ok && (roleStr == "vip" || roleStr == "admin") {
		resp["subscription"] = "vip"     // 标记为 VIP 订阅
		resp["rate_limit"] = 1000        // VIP 速率限制：1000 次/分钟
		resp["vip_remaining"] = 10000    // VIP 剩余调用次数
	}

	// 步骤4：返回统计数据
	response.OK(c, resp)
}
