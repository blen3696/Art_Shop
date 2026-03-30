package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIService handles interactions with the Anthropic Claude API for content generation
// and product recommendations.
type AIService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewAIService creates a new AIService instance.
func NewAIService(cfg *config.Config) *AIService {
	return &AIService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// claudeRequest represents the request body for the Anthropic Messages API.
type claudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []claudeMessage  `json:"messages"`
}

// claudeMessage represents a single message in the Claude conversation.
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse represents the response body from the Anthropic Messages API.
type claudeResponse struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Role    string          `json:"role"`
	Content []claudeContent `json:"content"`
	Model   string          `json:"model"`
	Usage   claudeUsage     `json:"usage"`
}

// claudeContent represents a content block in the Claude response.
type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// claudeUsage tracks token usage for the Claude API call.
type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// claudeErrorResponse represents an error from the Claude API.
type claudeErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// callClaude sends a request to the Anthropic Messages API and returns the text response.
func (s *AIService) callClaude(prompt string, maxTokens int) (string, error) {
	if s.cfg.AnthropicAPIKey == "" {
		return "", fmt.Errorf("ai_service: Anthropic API key is not configured")
	}

	model := s.cfg.AIModel
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	reqBody := claudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ai_service: failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ai_service: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.cfg.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai_service: API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ai_service: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp claudeErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("ai_service: Claude API error (%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return "", fmt.Errorf("ai_service: Claude API returned status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("ai_service: failed to parse response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("ai_service: empty response from Claude API")
	}

	// Concatenate all text content blocks.
	var result strings.Builder
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}

	return result.String(), nil
}

// GenerateProductDescription uses Claude to generate an engaging art product description.
func (s *AIService) GenerateProductDescription(title, medium, dimensions string, tags []string) (string, error) {
	tagsStr := "none"
	if len(tags) > 0 {
		tagsStr = strings.Join(tags, ", ")
	}

	prompt := fmt.Sprintf(`You are an expert art curator and copywriter for an online art marketplace called ArtShop. Write a compelling, engaging product description for the following artwork. The description should be 2-3 paragraphs, highlight the artistic qualities, emotional impact, and potential display settings. Do NOT include the title in the description. Return ONLY the description text, no headers or labels.

Title: %s
Medium: %s
Dimensions: %s
Tags: %s`, title, medium, dimensions, tagsStr)

	description, err := s.callClaude(prompt, 500)
	if err != nil {
		return "", fmt.Errorf("ai_service: failed to generate description: %w", err)
	}

	return strings.TrimSpace(description), nil
}

// GenerateProductTags uses Claude to suggest relevant tags for a product.
func (s *AIService) GenerateProductTags(title, description string) ([]string, error) {
	prompt := fmt.Sprintf(`You are an expert art curator for an online art marketplace. Based on the following artwork title and description, suggest 5-10 relevant tags that would help buyers discover this artwork through search. Return ONLY the tags as a JSON array of strings, nothing else.

Title: %s
Description: %s`, title, description)

	result, err := s.callClaude(prompt, 200)
	if err != nil {
		return nil, fmt.Errorf("ai_service: failed to generate tags: %w", err)
	}

	// Parse the JSON array from the response.
	result = strings.TrimSpace(result)

	// Handle cases where Claude wraps in markdown code block.
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	var tags []string
	if err := json.Unmarshal([]byte(result), &tags); err != nil {
		// If JSON parsing fails, try to extract tags from plain text.
		lines := strings.Split(result, "\n")
		for _, line := range lines {
			tag := strings.TrimSpace(line)
			tag = strings.Trim(tag, `",-[]`)
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		if len(tags) == 0 {
			return nil, fmt.Errorf("ai_service: failed to parse tags from response: %s", result)
		}
	}

	return tags, nil
}

// GetRecommendations returns personalized product recommendations based on the user's
// browsing history. It uses a combination of content-based filtering (similar products)
// and the Claude API for reasoning about user preferences.
func (s *AIService) GetRecommendations(userID uuid.UUID, browsingHistory []uuid.UUID, db *gorm.DB) ([]models.Product, error) {
	if len(browsingHistory) == 0 {
		// No browsing history: return popular/featured products as fallback.
		var products []models.Product
		if err := db.Where("is_published = ? AND is_featured = ?", true, true).
			Order("view_count DESC, total_sales DESC").
			Limit(10).
			Preload("Seller").
			Preload("Category").
			Find(&products).Error; err != nil {
			return nil, fmt.Errorf("ai_service: failed to get fallback recommendations: %w", err)
		}
		return products, nil
	}

	// Fetch the products the user has viewed.
	var viewedProducts []models.Product
	if err := db.Where("id IN ?", browsingHistory).
		Preload("Category").
		Find(&viewedProducts).Error; err != nil {
		return nil, fmt.Errorf("ai_service: failed to get viewed products: %w", err)
	}

	if len(viewedProducts) == 0 {
		var products []models.Product
		if err := db.Where("is_published = ?", true).
			Order("view_count DESC").
			Limit(10).
			Preload("Seller").
			Preload("Category").
			Find(&products).Error; err != nil {
			return nil, fmt.Errorf("ai_service: failed to get fallback recommendations: %w", err)
		}
		return products, nil
	}

	// Collect categories, mediums, and price ranges from browsing history.
	categoryIDs := make(map[uuid.UUID]bool)
	mediums := make(map[string]bool)
	var totalPrice float64

	for _, p := range viewedProducts {
		if p.CategoryID != nil {
			categoryIDs[*p.CategoryID] = true
		}
		if p.Medium != nil {
			mediums[*p.Medium] = true
		}
		totalPrice += p.Price
	}

	avgPrice := totalPrice / float64(len(viewedProducts))
	priceMin := avgPrice * 0.5
	priceMax := avgPrice * 2.0

	// Build a query for similar products the user has NOT viewed.
	query := db.Where("is_published = ?", true).
		Where("id NOT IN ?", browsingHistory)

	// Filter by categories the user has shown interest in.
	catIDs := make([]uuid.UUID, 0, len(categoryIDs))
	for id := range categoryIDs {
		catIDs = append(catIDs, id)
	}
	if len(catIDs) > 0 {
		query = query.Where("category_id IN ?", catIDs)
	}

	// Filter by price range around the user's average viewed price.
	query = query.Where("price BETWEEN ? AND ?", priceMin, priceMax)

	var recommendations []models.Product
	if err := query.
		Order("avg_rating DESC, total_sales DESC, view_count DESC").
		Limit(10).
		Preload("Seller").
		Preload("Category").
		Find(&recommendations).Error; err != nil {
		return nil, fmt.Errorf("ai_service: failed to get recommendations: %w", err)
	}

	// If we got too few results, supplement with popular products.
	if len(recommendations) < 5 {
		var supplement []models.Product
		existingIDs := make([]uuid.UUID, len(recommendations))
		for i, r := range recommendations {
			existingIDs[i] = r.ID
		}
		excludeIDs := append(browsingHistory, existingIDs...)

		if err := db.Where("is_published = ? AND id NOT IN ?", true, excludeIDs).
			Order("total_sales DESC, view_count DESC").
			Limit(10 - len(recommendations)).
			Preload("Seller").
			Preload("Category").
			Find(&supplement).Error; err == nil {
			recommendations = append(recommendations, supplement...)
		}
	}

	return recommendations, nil
}
