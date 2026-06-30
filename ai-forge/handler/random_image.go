package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
)

// RandomImageRequest represents the request parameters for random image
type RandomImageRequest struct {
	Width    int    `form:"width" json:"width"`
	Height   int    `form:"height" json:"height"`
	Category string `form:"category" json:"category"`
	Format   string `form:"format" json:"format"` // "url" or "redirect"
}

// RandomImageResponse represents the response for random image
type RandomImageResponse struct {
	URL        string `json:"url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Category   string `json:"category,omitempty"`
	Provider   string `json:"provider"`
	GeneratedAt string `json:"generated_at"`
}

// ImageCategory represents supported image categories
var ImageCategory = map[string]string{
	"nature":      "nature",
	"architecture": "architecture",
	"people":       "people",
	"technology":   "technology",
	"animals":      "animals",
	"food":         "food",
	"travel":       "travel",
	"abstract":     "abstract",
}

// GetRandomImage returns a random image URL
func GetRandomImage(c *gin.Context) {
	var req RandomImageRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	// Set defaults
	if req.Width <= 0 {
		req.Width = 1280
	}
	if req.Height <= 0 {
		req.Height = 720
	}
	if req.Format == "" {
		req.Format = "url"
	}

	// Validate dimensions
	if req.Width < 100 || req.Width > 4096 {
		response.BadRequest(c, "width must be between 100 and 4096")
		return
	}
	if req.Height < 100 || req.Height > 4096 {
		response.BadRequest(c, "height must be between 100 and 4096")
		return
	}

	// Validate category if provided
	if req.Category != "" {
		if _, exists := ImageCategory[strings.ToLower(req.Category)]; !exists {
			validCategories := make([]string, 0, len(ImageCategory))
			for k := range ImageCategory {
				validCategories = append(validCategories, k)
			}
			response.BadRequest(c, fmt.Sprintf("invalid category, valid options: %s", strings.Join(validCategories, ", ")))
			return
		}
	}

	// Generate random image URL
	imageURL := generateRandomImageURL(req.Width, req.Height, req.Category)

	// Build response
	resp := RandomImageResponse{
		URL:         imageURL,
		Width:       req.Width,
		Height:      req.Height,
		Category:    req.Category,
		Provider:    "picsum.photos",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// If format is redirect, redirect to the image URL
	if req.Format == "redirect" {
		c.Redirect(http.StatusTemporaryRedirect, imageURL)
		return
	}

	// Otherwise return JSON response
	response.OK(c, resp)
}

// generateRandomImageURL generates a random image URL from Picsum
func generateRandomImageURL(width, height int, category string) string {
	// Use Picsum for high-quality placeholder images
	// Format: https://picsum.photos/{width}/{height}
	// With random seed: https://picsum.photos/{width}/{height}?random={seed}

	// Generate random seed for variety
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	seed := r.Intn(10000)

	// Build base URL
	baseURL := fmt.Sprintf("https://picsum.photos/%d/%d", width, height)

	// Add random seed to ensure different images
	params := url.Values{}
	params.Add("random", strconv.Itoa(seed))

	// Add grayscale if category is abstract
	if strings.ToLower(category) == "abstract" {
		params.Add("grayscale", "")
	}

	// Add blur if category is abstract
	if strings.ToLower(category) == "abstract" {
		params.Add("blur", "")
	}

	if len(params) > 0 {
		return baseURL + "?" + params.Encode()
	}

	return baseURL
}

// GetRandomImageRedirect redirects to a random image
func GetRandomImageRedirect(c *gin.Context) {
	// Get dimensions from query params
	width, _ := strconv.Atoi(c.DefaultQuery("width", "1280"))
	height, _ := strconv.Atoi(c.DefaultQuery("height", "720"))
	category := c.Query("category")

	// Generate URL
	imageURL := generateRandomImageURL(width, height, category)

	// Redirect to the image
	c.Redirect(http.StatusTemporaryRedirect, imageURL)
}
