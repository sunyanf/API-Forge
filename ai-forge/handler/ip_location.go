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

// IPLocationRequest represents the request for IP location lookup
type IPLocationRequest struct {
	IP string `form:"ip" json:"ip"`
}

// IPLocationResponse represents the response for IP location lookup
type IPLocationResponse struct {
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
	RegionName  string `json:"region_name"`
	City        string `json:"city"`
	Zip         string `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string `json:"timezone"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	AS          string `json:"as"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

// externalIPAPIResponse represents the external API response
type externalIPAPIResponse struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Message     string  `json:"message,omitempty"`
}

// GetIPLocation handles IP geolocation lookup
func GetIPLocation(c *gin.Context) {
	var req IPLocationRequest

	// Bind query parameter or JSON body
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	// If no IP provided, use client's IP
	if req.IP == "" {
		req.IP = c.ClientIP()
	}

	// Validate IP format
	if net.ParseIP(req.IP) == nil {
		response.BadRequest(c, "invalid IP address format")
		return
	}

	// Check if it's a private IP
	if isPrivateIP(req.IP) {
		response.BadRequest(c, "cannot lookup private IP address")
		return
	}

	// Call external IP geolocation API
	location, err := lookupIPLocation(req.IP)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("failed to lookup IP location: %v", err))
		return
	}

	// Check if the lookup was successful
	if location.Status != "success" {
		response.BadRequest(c, fmt.Sprintf("IP lookup failed: %s", location.Message))
		return
	}

	// Map to our response format
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
		Status:      "success",
	}

	response.OK(c, resp)
}

// lookupIPLocation calls the external IP geolocation API
func lookupIPLocation(ip string) (*externalIPAPIResponse, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result externalIPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// isPrivateIP checks if an IP address is private
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	privateRanges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
		{net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")},
	}

	ipBytes := ip.To4()
	if ipBytes == nil {
		return false // IPv6 not handled here
	}

	for _, r := range privateRanges {
		startBytes := r.start.To4()
		endBytes := r.end.To4()
		if startBytes == nil || endBytes == nil {
			continue
		}

		if compareIP(ipBytes, startBytes) >= 0 && compareIP(ipBytes, endBytes) <= 0 {
			return true
		}
	}

	return false
}

// compareIP compares two IP addresses
func compareIP(a, b net.IP) int {
	for i := range a {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
