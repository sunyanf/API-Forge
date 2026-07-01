// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现随机图片（Random Image）的请求处理逻辑。
// 使用 picsum.photos 作为图源，支持指定宽高、分类和返回格式。
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
)

// RandomImageRequest 定义随机图片请求的参数
// 支持通过 URL 查询参数传递，客户端可指定图片的尺寸、分类和返回格式
type RandomImageRequest struct {
	// Width 图片宽度（像素），范围 100-4096，默认 1280
	Width int `form:"width" json:"width"`
	// Height 图片高度（像素），范围 100-4096，默认 720
	Height int `form:"height" json:"height"`
	// Category 图片分类，可选值见 ImageCategory，如 nature、architecture 等
	Category string `form:"category" json:"category"`
	// Format 返回格式："url" 表示返回 JSON（含图片 URL），"redirect" 表示直接重定向到图片
	Format string `form:"format" json:"format"`
}

// RandomImageResponse 定义随机图片的响应结构
type RandomImageResponse struct {
	// URL 生成的图片地址
	URL string `json:"url"`
	// Width 图片宽度（像素）
	Width int `json:"width"`
	// Height 图片高度（像素）
	Height int `json:"height"`
	// Category 图片分类，为空时不返回此字段
	Category string `json:"category,omitempty"`
	// Provider 图片来源服务商
	Provider string `json:"provider"`
	// GeneratedAt 图片生成时间，RFC3339 格式
	GeneratedAt string `json:"generated_at"`
}

// ImageCategory 定义支持的图片分类及其标识符
// 键为客户端传入的分类名，值为内部使用的分类标识
var ImageCategory = map[string]string{
	"nature":       "nature",       // 自然风景
	"architecture": "architecture", // 建筑
	"people":       "people",       // 人物
	"technology":   "technology",   // 科技
	"animals":      "animals",      // 动物
	"food":         "food",         // 美食
	"travel":       "travel",       // 旅行
	"abstract":     "abstract",     // 抽象/艺术
}

// GetRandomImage 处理随机图片获取请求
// 支持两种返回模式：
//  1. URL 模式（默认）：返回 JSON，包含图片地址和元信息
//  2. Redirect 模式：直接将客户端重定向到图片地址
//
// 请求示例（URL 模式）：
//
//	GET /api/random-image?width=1920&height=1080&category=nature
//
// 请求示例（Redirect 模式）：
//
//	GET /api/random-image?width=800&height=600&format=redirect
func GetRandomImage(c *gin.Context) {
	var req RandomImageRequest

	// 步骤1：绑定 URL 查询参数到请求结构体
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	// 步骤2：设置默认值
	// 如果客户端未提供宽度或提供了无效值（<=0），使用默认值
	if req.Width <= 0 {
		req.Width = 1280 // 默认宽度 1280px
	}
	if req.Height <= 0 {
		req.Height = 720 // 默认高度 720px
	}
	if req.Format == "" {
		req.Format = "url" // 默认返回 JSON 格式
	}

	// 步骤3：验证图片尺寸范围
	// 防止请求过大的图片导致图源服务器压力过大
	if req.Width < 100 || req.Width > 4096 {
		response.BadRequest(c, "width must be between 100 and 4096")
		return
	}
	if req.Height < 100 || req.Height > 4096 {
		response.BadRequest(c, "height must be between 100 and 4096")
		return
	}

	// 步骤4：验证分类参数（如果客户端提供了）
	if req.Category != "" {
		// 转为小写进行不区分大小写的匹配
		if _, exists := ImageCategory[strings.ToLower(req.Category)]; !exists {
			// 分类不存在，收集合法选项用于错误提示
			validCategories := make([]string, 0, len(ImageCategory))
			for k := range ImageCategory {
				validCategories = append(validCategories, k)
			}
			response.BadRequest(c, fmt.Sprintf("invalid category, valid options: %s", strings.Join(validCategories, ", ")))
			return
		}
	}

	// 步骤5：调用图片 URL 生成函数
	imageURL := generateRandomImageURL(req.Width, req.Height, req.Category)

	// 步骤6：构造响应体
	resp := RandomImageResponse{
		URL:         imageURL,                                 // 生成的图片 URL
		Width:       req.Width,                                // 图片宽度
		Height:      req.Height,                               // 图片高度
		Category:    req.Category,                             // 图片分类（可能为空）
		Provider:    "loremflickr.com",                          // 图片来源
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),    // 生成时间
	}

	// 步骤7：根据 format 参数决定返回方式
	if req.Format == "redirect" {
		// Redirect 模式：浏览器会直接跳转到图片地址
		// 状态码 307 Temporary Redirect 表示临时重定向
		c.Redirect(http.StatusTemporaryRedirect, imageURL)
		return
	}

	// URL 模式（默认）：返回包含图片 URL 的 JSON 响应
	response.OK(c, resp)
}

// generateRandomImageURL 生成一个随机的 picsum.photos 图片 URL
//
// 参数：
//   - width: 图片宽度（像素）
//   - height: 图片高度（像素）
//   - category: 图片分类，抽象类图片会添加灰度+模糊效果
//
// 返回值：
//   - string: 可直接访问的图片 URL
//
// 实现原理：
// 每次调用生成不同的 random seed 参数，确保同一尺寸也能获取不同图片。
// 对于 abstract 分类，额外添加 grayscale（灰度）和 blur（模糊）参数。
// generateRandomImageURL 根据分类生成图片 URL
// 使用 loremflickr.com——支持按类别（人群、自然、建筑等）返回对应图片
func generateRandomImageURL(width, height int, category string) string {
	cat := strings.ToLower(category)

	// loremflickr.com 支持的真实分类关键词
	loremCat := ""
	switch cat {
	case "nature":
		loremCat = "nature,landscape"
	case "architecture":
		loremCat = "architecture,building"
	case "people":
		loremCat = "people,portrait"
	case "technology":
		loremCat = "computer,technology"
	case "animals":
		loremCat = "animals,wildlife"
	case "food":
		loremCat = "food,meal"
	case "travel":
		loremCat = "travel,vacation"
	case "abstract":
		loremCat = "abstract,art"
	}

	if loremCat != "" {
		// loremflickr 按关键词返回对应类别的图片
		return fmt.Sprintf("https://loremflickr.com/%d/%d/%s?random=%d", width, height, loremCat, time.Now().UnixNano()%100000)
	}
	// 没选类别时用 picsum 通用随机
	return fmt.Sprintf("https://picsum.photos/seed/%d/%d/%d", time.Now().UnixNano()%100000, width, height)
}

// GetRandomImageRedirect 以重定向方式返回随机图片
// 这是一个便捷接口，等于调用 GetRandomImage 并将 format 固定为 redirect。
// 客户端访问此接口后会直接看到图片，不会得到 JSON 响应。
//
// 请求示例：
//
//	GET /api/random-image/redirect?width=800&height=600&category=nature
func GetRandomImageRedirect(c *gin.Context) {
	// 步骤1：从查询参数获取尺寸，使用默认值兜底
	// DefaultQuery 在参数不存在时返回默认值
	width, _ := strconv.Atoi(c.DefaultQuery("width", "1280"))
	height, _ := strconv.Atoi(c.DefaultQuery("height", "720"))
	category := c.Query("category")

	// 步骤2：生成随机图片 URL
	imageURL := generateRandomImageURL(width, height, category)

	// 步骤3：302 临时重定向到图片地址
	// 浏览器会自动跳转并显示图片
	c.Redirect(http.StatusTemporaryRedirect, imageURL)
}
