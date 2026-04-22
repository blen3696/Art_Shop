package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrAIDisabled is returned when GEMINI_API_KEY is not configured. Handlers
// should translate this into HTTP 503 Service Unavailable so the UI can
// gracefully hide AI-powered affordances without surfacing a 500.
var ErrAIDisabled = errors.New("ai_service: AI is not configured")

const defaultGeminiModel = "gemini-2.0-flash"
// Google renamed text-embedding-004 → gemini-embedding-001. The new model
// defaults to 3072 dims but supports 768/1536/3072 via outputDimensionality.
// We keep 768 so existing vectors in the DB stay compatible.
const defaultGeminiEmbedModel = "gemini-embedding-001"
const embeddingDimension = 768
const geminiEndpointTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
const geminiEmbedEndpointTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent"

// AIService handles AI-powered content generation (via Google's Gemini API)
// and product recommendations (pure SQL collaborative filtering — works
// regardless of whether an AI key is configured).
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

// IsEnabled reports whether generative-AI features are available.
// Recommendations don't require this (they're SQL-based).
func (s *AIService) IsEnabled() bool {
	return s.cfg.GeminiAPIKey != ""
}

// --- Gemini REST wire types -------------------------------------------------

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content      geminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// callGemini sends a prompt to Google's Gemini API and returns the text response.
// Returns ErrAIDisabled (wrapped) when no API key is configured.
func (s *AIService) callGemini(prompt string, maxTokens int) (string, error) {
	if !s.IsEnabled() {
		return "", ErrAIDisabled
	}

	model := s.cfg.AIModel
	if model == "" {
		model = defaultGeminiModel
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{{
			Parts: []geminiPart{{Text: prompt}},
		}},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens: maxTokens,
			Temperature:     0.7,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ai_service: marshal request: %w", err)
	}

	endpoint := fmt.Sprintf(geminiEndpointTemplate, url.PathEscape(model)) +
		"?key=" + url.QueryEscape(s.cfg.GeminiAPIKey)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ai_service: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai_service: API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ai_service: read response: %w", err)
	}

	var parsed geminiResponse
	_ = json.Unmarshal(body, &parsed)

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return "", fmt.Errorf("ai_service: Gemini API error (%d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("ai_service: Gemini API returned status %d", resp.StatusCode)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("ai_service: empty response from Gemini API")
	}

	var result strings.Builder
	for _, p := range parsed.Candidates[0].Content.Parts {
		result.WriteString(p.Text)
	}
	return result.String(), nil
}

// --- Embeddings -------------------------------------------------------------

type geminiEmbedRequest struct {
	Model                string        `json:"model"`
	Content              geminiContent `json:"content"`
	OutputDimensionality int           `json:"outputDimensionality,omitempty"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed returns a 768-dimension vector for the given text using Gemini's
// gemini-embedding-001 model. Returns ErrAIDisabled if no key is configured.
func (s *AIService) Embed(text string) ([]float32, error) {
	if !s.IsEnabled() {
		return nil, ErrAIDisabled
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("ai_service: cannot embed empty text")
	}

	reqBody := geminiEmbedRequest{
		Model: "models/" + defaultGeminiEmbedModel,
		Content: geminiContent{
			Parts: []geminiPart{{Text: trimmed}},
		},
		OutputDimensionality: embeddingDimension,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ai_service: marshal embed request: %w", err)
	}

	endpoint := fmt.Sprintf(geminiEmbedEndpointTemplate, url.PathEscape(defaultGeminiEmbedModel)) +
		"?key=" + url.QueryEscape(s.cfg.GeminiAPIKey)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("ai_service: build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai_service: embed request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ai_service: read embed response: %w", err)
	}

	var parsed geminiEmbedResponse
	_ = json.Unmarshal(body, &parsed)

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return nil, fmt.Errorf("ai_service: Gemini embed error (%d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return nil, fmt.Errorf("ai_service: Gemini embed returned status %d", resp.StatusCode)
	}

	if len(parsed.Embedding.Values) != embeddingDimension {
		return nil, fmt.Errorf("ai_service: unexpected embedding dimension: got %d, want %d",
			len(parsed.Embedding.Values), embeddingDimension)
	}

	return parsed.Embedding.Values, nil
}

// FormatVector formats a float slice as a pgvector literal: "[0.1,0.2,...]".
// Use this when building parameters for ::vector casts in raw SQL.
func FormatVector(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.Grow(len(v) * 10)
	b.WriteByte('[')
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%g", x)
	}
	b.WriteByte(']')
	return b.String()
}

// BuildProductEmbeddingText assembles the text that represents a product for
// embedding. Exposed so other services can detect when re-embedding is needed.
func BuildProductEmbeddingText(title, description, medium string, tags []string) string {
	parts := []string{title}
	if description != "" {
		parts = append(parts, description)
	}
	if medium != "" {
		parts = append(parts, "Medium: "+medium)
	}
	if len(tags) > 0 {
		parts = append(parts, "Tags: "+strings.Join(tags, ", "))
	}
	return strings.Join(parts, "\n")
}

// GenerateProductDescription uses Gemini to generate an engaging art product description.
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

	description, err := s.callGemini(prompt, 500)
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

	result, err := s.callGemini(prompt, 200)
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
