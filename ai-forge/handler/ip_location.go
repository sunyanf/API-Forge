// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现 IP 归属地查询（IP Location）的请求处理逻辑。
// 通过调用外部 API（ip-api.com）查询指定 IP 的地理位置信息，
// 包括国家、地区、城市、经纬度、运营商等。
package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
)

// IPLocationRequest 定义 IP 归属地查询的请求参数
// 支持通过查询参数（form）或 JSON 请求体传递 IP 地址
type IPLocationRequest struct {
	// IP 要查询的 IP 地址，支持 form 和 json 两种格式绑定
	IP string `form:"ip" json:"ip"`
}

// IPLocationResponse 定义 IP 归属地查询的响应结构
// 包含 IP 地址的地理位置详细信息
type IPLocationResponse struct {
	// IP 查询的 IP 地址
	IP string `json:"ip"`
	// Country 所属国家名称（中文）
	Country string `json:"country"`
	// CountryCode 国家代码（ISO 3166-1 alpha-2），如 CN、US
	CountryCode string `json:"country_code"`
	// Region 地区缩写代码，如 BJ（北京）、GD（广东）
	Region string `json:"region"`
	// RegionName 地区全名（中文）
	RegionName string `json:"region_name"`
	// City 城市名称（中文）
	City string `json:"city"`
	// Zip 邮政编码
	Zip string `json:"zip"`
	// Lat 纬度坐标
	Lat float64 `json:"lat"`
	// Lon 经度坐标
	Lon float64 `json:"lon"`
	// Timezone 所在时区，如 Asia/Shanghai
	Timezone string `json:"timezone"`
	// ISP 互联网服务提供商名称
	ISP string `json:"isp"`
	// Org 所属组织机构
	Org string `json:"org"`
	// AS 自治系统号及名称
	AS string `json:"as"`
	// Status 查询状态，success 表示成功
	Status string `json:"status"`
	// Message 当查询失败时返回的错误信息，omitempty 表示成功时不显示此字段
	Message string `json:"message,omitempty"`
}

// externalIPAPIResponse 定义外部 IP 查询 API（ip-api.com）的响应结构
// 用于反序列化外部 API 返回的 JSON 数据，字段名与外部 API 保持一致
type externalIPAPIResponse struct {
	// Query 被查询的 IP 地址
	Query string `json:"query"`
	// Status 查询状态，success 或 fail
	Status string `json:"status"`
	// Country 国家名称
	Country string `json:"country"`
	// CountryCode 国家代码
	CountryCode string `json:"countryCode"`
	// Region 地区代码
	Region string `json:"region"`
	// RegionName 地区名称
	RegionName string `json:"regionName"`
	// City 城市名称
	City string `json:"city"`
	// Zip 邮编
	Zip string `json:"zip"`
	// Lat 纬度
	Lat float64 `json:"lat"`
	// Lon 经度
	Lon float64 `json:"lon"`
	// Timezone 时区
	Timezone string `json:"timezone"`
	// ISP 运营商
	ISP string `json:"isp"`
	// Org 组织
	Org string `json:"org"`
	// AS AS 号
	AS string `json:"as"`
	// Message 错误消息，查询成功时为空
	Message string `json:"message,omitempty"`
}

// GetIPLocation 处理 IP 归属地查询请求
// 接收一个 IP 地址（可选，不传则使用客户端 IP），调用 ip-api.com 查询其地理位置信息。
// 查询前会验证 IP 格式并拒绝内网地址。
//
// 请求示例：
//
//	GET /api/ip-location?ip=8.8.8.8
//
// 响应（成功，200 OK）：
//
//	{"ip": "8.8.8.8", "country": "美国", "city": "Mountain View", ...}
func GetIPLocation(c *gin.Context) {
	var req IPLocationRequest

	// 步骤1：绑定查询参数（支持 URL query string 和 JSON body）
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	// 步骤2：如果客户端未提供 IP 地址，使用 ip-api.com 自动检测请求者的公网 IP
	// 注意：传空字符串给 ip-api.com 会自动使用请求来源的公网 IP
	if req.IP == "" {
		// 使用 c.ClientIP() 获取真实客户端 IP（经过代理时取 X-Forwarded-For）
		clientIP := c.ClientIP()
		if clientIP == "::1" || clientIP == "127.0.0.1" || isPrivateIP(clientIP) {
			// 本地开发环境：让 ip-api.com 自动检测（传空或传自身公网 IP 均可）
			// 传空字符串给外部 API 时，ip-api 会用请求来源 IP 查询
			req.IP = "" // 保持空，后续 lookupIPLocation 会处理
		} else {
			req.IP = clientIP
		}
	}

	// 步骤3：验证 IP 地址格式（如果用户提供了 IP 的话）
	if req.IP != "" && net.ParseIP(req.IP) == nil {
		response.BadRequest(c, "IP 地址格式无效")
		return
	}

	// 步骤4：检查是否为内网 IP（仅当用户提供了 IP 时）
	if req.IP != "" && isPrivateIP(req.IP) {
		response.BadRequest(c, "无法查询内网 IP 地址")
		return
	}

	// 步骤5：调用外部 IP 地理位置 API 进行查询
	location, err := lookupIPLocation(req.IP)
	if err != nil {
		// 网络请求失败（超时、DNS 解析失败等），返回 500 错误
		response.InternalError(c, fmt.Sprintf("IP 查询失败: %v", err))
		return
	}

	// 步骤6：检查外部 API 的查询结果状态
	// 即使 HTTP 请求成功，API 也可能返回业务逻辑上的失败（如 IP 地址无效）
	if location.Status != "success" {
		response.BadRequest(c, fmt.Sprintf("IP 查询失败: %s", location.Message))
		return
	}

	// 步骤7：将外部 API 的响应字段映射到统一的内部响应格式
	// 这样做的好处是：如果将来更换外部 API 提供商，只需修改映射逻辑
	resp := IPLocationResponse{
		IP:          location.Query,
		Country:     location.Country,
		CountryCode: location.CountryCode,
		Region:      location.Region,
		RegionName:  location.RegionName,
		City:        location.City,
		Zip:         location.Zip,
		Lat:         location.Lat,
		Lon:         location.Lon,
		Timezone:    location.Timezone,
		ISP:         location.ISP,
		Org:         location.Org,
		AS:          location.AS,
		Status:      "success", // 统一标记为成功
	}

	// 步骤8：返回 200 OK 及地理位置信息
	response.OK(c, resp)
}

// lookupIPLocation 调用 ip-api.com 的外部 API 查询指定 IP 的地理位置
// 参数 ip: 要查询的 IP 地址，空字符串表示查询请求来源的 IP
func lookupIPLocation(ip string) (*externalIPAPIResponse, error) {
	// 构造请求 URL——不传 IP 时 ip-api.com 会自动检测请求来源的 IP
	urlStr := "http://ip-api.com/json/?lang=zh-CN"
	if ip != "" {
		urlStr = fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip)
	}

	// 创建 HTTP 客户端，设置 10 秒超时，避免因外部 API 不可用导致请求阻塞
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送 GET 请求
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err // 网络层面的错误（DNS 失败、连接超时等）
	}
	defer resp.Body.Close() // 确保响应体最终被关闭，防止资源泄漏

	// 将响应体的 JSON 数据解码到结构体中
	var result externalIPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err // JSON 解析失败
	}

	return &result, nil
}

// isPrivateIP 检查指定的 IP 地址是否属于私有/内网地址段
//
// 参数：
//   - ipStr: 待检查的 IP 地址字符串
//
// 返回值：
//   - bool: true 表示是内网地址，false 表示是公网地址
//
// 内网地址段包括：
//   - 10.0.0.0/8       （A 类私有地址）
//   - 172.16.0.0/12    （B 类私有地址）
//   - 192.168.0.0/16   （C 类私有地址）
//   - 127.0.0.0/8      （本地回环地址）
func isPrivateIP(ipStr string) bool {
	// 先将字符串解析为 net.IP 类型
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false // 无法解析的 IP 不认为是内网地址
	}

	// 定义四个常见的内网地址段，使用起止 IP 表示范围
	privateRanges := []struct {
		start net.IP // 网段起始地址
		end   net.IP // 网段结束地址
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},       // A 类私有地址
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},     // B 类私有地址
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},   // C 类私有地址
		{net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")},     // 回环地址
	}

	// 将 IP 转换为 4 字节的 IPv4 表示
	ipBytes := ip.To4()
	if ipBytes == nil {
		return false // IPv6 地址暂不在此处理
	}

	// 遍历每个内网地址段，判断目标 IP 是否落在其中
	for _, r := range privateRanges {
		// 将网段的起止 IP 也转为 4 字节表示
		startBytes := r.start.To4()
		endBytes := r.end.To4()
		if startBytes == nil || endBytes == nil {
			continue // 转换失败则跳过此范围
		}

		// 使用字节比较判断 IP 是否在 [start, end] 范围内
		// compareIP 返回值 >= 0 表示 ipBytes >= startBytes
		// compareIP 返回值 <= 0 表示 ipBytes <= endBytes
		if compareIP(ipBytes, startBytes) >= 0 && compareIP(ipBytes, endBytes) <= 0 {
			return true // IP 落在内网地址段中
		}
	}

	return false // IP 不在任何内网地址段中，是公网 IP
}

// compareIP 按字节比较两个 IP 地址的大小（用于范围判断）
//
// 参数：
//   - a: 第一个 IP 的字节表示
//   - b: 第二个 IP 的字节表示
//
// 返回值：
//   - int: 如果 a < b 返回 -1，a == b 返回 0，a > b 返回 1
//
// 比较逻辑：从左到右逐字节比较，找到第一个不相等的字节即返回结果
func compareIP(a, b net.IP) int {
	for i := range a {
		if a[i] < b[i] {
			return -1 // a 当前字节小于 b，所以 a < b
		}
		if a[i] > b[i] {
			return 1 // a 当前字节大于 b，所以 a > b
		}
	}
	return 0 // 所有字节相等，a == b
}
