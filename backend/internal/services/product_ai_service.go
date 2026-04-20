package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Lifetime of a cached review summary. Summaries are also re-generated sooner
// when the review count changes.
const reviewSummaryMaxAge = 7 * 24 * time.Hour

// ProductAIService owns all AI-powered product features: semantic search,
// "similar products", embedding storage, and review summaries. All methods
// degrade gracefully when the underlying AI service is disabled.
type ProductAIService struct {
	productRepo *repository.ProductRepository
	reviewRepo  *repository.ReviewRepository
	ai          *AIService
	db          *gorm.DB
}

func NewProductAIService(
	productRepo *repository.ProductRepository,
	reviewRepo *repository.ReviewRepository,
	ai *AIService,
	db *gorm.DB,
) *ProductAIService {
	return &ProductAIService{
		productRepo: productRepo,
		reviewRepo:  reviewRepo,
		ai:          ai,
		db:          db,
	}
}

// --- Semantic search --------------------------------------------------------

// SearchWithFallback returns products matching a query. When AI is enabled,
// uses vector similarity. Otherwise (or on failure), falls back to the
// keyword search so results are always returned.
func (s *ProductAIService) SearchWithFallback(query string, page, perPage int) ([]models.Product, int64, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, 0, fmt.Errorf("product_ai: empty query")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	if s.ai.IsEnabled() {
		products, total, err := s.searchSemantic(query, page, perPage)
		if err == nil {
			return products, total, nil
		}
		slog.Warn("product_ai: semantic search failed, falling back to keyword", "error", err)
	}
	return s.productRepo.Search(query, page, perPage)
}

func (s *ProductAIService) searchSemantic(query string, page, perPage int) ([]models.Product, int64, error) {
	vec, err := s.ai.Embed(query)
	if err != nil {
		return nil, 0, err
	}
	vecLit := FormatVector(vec)

	// Count total rows eligible for semantic ranking (i.e. those that have an embedding).
	var total int64
	if err := s.db.Model(&models.Product{}).
		Where("is_published = ? AND embedding IS NOT NULL", true).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("product_ai: count: %w", err)
	}
	if total == 0 {
		return nil, 0, errors.New("product_ai: no embedded products yet")
	}

	offset := (page - 1) * perPage
	var ids []uuid.UUID
	if err := s.db.Raw(`
		SELECT id FROM products
		WHERE is_published = TRUE AND embedding IS NOT NULL
		ORDER BY embedding <=> ?::vector
		LIMIT ? OFFSET ?`,
		vecLit, perPage, offset,
	).Scan(&ids).Error; err != nil {
		return nil, 0, fmt.Errorf("product_ai: vector search: %w", err)
	}

	if len(ids) == 0 {
		return []models.Product{}, total, nil
	}
	return s.loadProductsInIDOrder(ids, total)
}

func (s *ProductAIService) loadProductsInIDOrder(ids []uuid.UUID, total int64) ([]models.Product, int64, error) {
	var products []models.Product
	if err := s.db.
		Where("id IN ?", ids).
		Preload("Seller").
		Preload("Category").
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("product_ai: hydrate: %w", err)
	}

	// Re-order to match the similarity ranking from the raw query.
	order := make(map[uuid.UUID]int, len(ids))
	for i, id := range ids {
		order[id] = i
	}
	sorted := make([]models.Product, len(products))
	for _, p := range products {
		if idx, ok := order[p.ID]; ok && idx < len(sorted) {
			sorted[idx] = p
		}
	}
	// Drop any zero-value slots (product rows that didn't hydrate for any reason).
	out := sorted[:0]
	for _, p := range sorted {
		if p.ID != uuid.Nil {
			out = append(out, p)
		}
	}
	return out, total, nil
}

// --- Similar products -------------------------------------------------------

// Similar returns up to `limit` products most similar to the given product by
// embedding cosine distance. Returns an empty slice when AI is disabled or
// the product has no embedding yet.
func (s *ProductAIService) Similar(productID uuid.UUID, limit int) ([]models.Product, error) {
	if limit < 1 || limit > 20 {
		limit = 4
	}
	if !s.ai.IsEnabled() {
		return s.similarFallback(productID, limit)
	}

	var ids []uuid.UUID
	err := s.db.Raw(`
		SELECT p.id
		FROM products p, products src
		WHERE src.id = ?
		  AND p.id <> src.id
		  AND p.is_published = TRUE
		  AND p.embedding IS NOT NULL
		  AND src.embedding IS NOT NULL
		ORDER BY p.embedding <=> src.embedding
		LIMIT ?`,
		productID, limit,
	).Scan(&ids).Error
	if err != nil {
		return nil, fmt.Errorf("product_ai: similar: %w", err)
	}

	if len(ids) == 0 {
		return s.similarFallback(productID, limit)
	}
	products, _, err := s.loadProductsInIDOrder(ids, int64(len(ids)))
	return products, err
}

// similarFallback returns other products in the same category as a reasonable
// approximation when embeddings aren't available.
func (s *ProductAIService) similarFallback(productID uuid.UUID, limit int) ([]models.Product, error) {
	src, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, err
	}

	q := s.db.
		Where("is_published = ?", true).
		Where("id <> ?", productID)
	if src.CategoryID != nil {
		q = q.Where("category_id = ?", *src.CategoryID)
	}

	var products []models.Product
	if err := q.Preload("Seller").Preload("Category").
		Order("total_sales DESC, view_count DESC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// --- Embedding management ---------------------------------------------------

// EmbedProduct generates + stores an embedding for a single product. Safe to
// call when AI is disabled — it returns nil without doing anything.
func (s *ProductAIService) EmbedProduct(p *models.Product) error {
	if !s.ai.IsEnabled() || p == nil {
		return nil
	}
	description := ""
	if p.Description != nil {
		description = *p.Description
	}
	medium := ""
	if p.Medium != nil {
		medium = *p.Medium
	}
	text := BuildProductEmbeddingText(p.Title, description, medium, []string(p.Tags))
	if strings.TrimSpace(text) == "" {
		return nil
	}

	vec, err := s.ai.Embed(text)
	if err != nil {
		return err
	}

	return s.db.Exec(`
		UPDATE products
		SET embedding = ?::vector, embedded_at = NOW(), embedding_src = ?
		WHERE id = ?`,
		FormatVector(vec), text, p.ID,
	).Error
}

// EmbedProductAsync runs EmbedProduct in a goroutine, logging any errors.
// Callers use this after Create/Update so the HTTP response isn't delayed.
func (s *ProductAIService) EmbedProductAsync(p *models.Product) {
	if p == nil || !s.ai.IsEnabled() {
		return
	}
	go func() {
		if err := s.EmbedProduct(p); err != nil {
			slog.Warn("product_ai: embed failed", "product_id", p.ID, "error", err)
		}
	}()
}

// BackfillEmbeddings embeds every published product that doesn't yet have an
// embedding. Respects Gemini's free-tier limits (~15 RPM) by sleeping between
// calls. Designed to run once at boot in a background goroutine.
func (s *ProductAIService) BackfillEmbeddings(ctx context.Context) {
	if !s.ai.IsEnabled() {
		return
	}
	const delay = 4500 * time.Millisecond // ~13 RPM, comfortably under the 15 RPM free tier cap

	for {
		if ctx.Err() != nil {
			return
		}
		var batch []models.Product
		if err := s.db.
			Where("is_published = ? AND embedding IS NULL", true).
			Limit(20).
			Find(&batch).Error; err != nil {
			slog.Warn("product_ai: backfill query failed", "error", err)
			return
		}
		if len(batch) == 0 {
			slog.Info("product_ai: embedding backfill complete")
			return
		}
		for i := range batch {
			if ctx.Err() != nil {
				return
			}
			if err := s.EmbedProduct(&batch[i]); err != nil {
				slog.Warn("product_ai: backfill embed failed",
					"product_id", batch[i].ID, "error", err)
			}
			time.Sleep(delay)
		}
	}
}

// --- Review summarisation ---------------------------------------------------

// ReviewSummary returns the cached LLM summary for a product's reviews,
// regenerating it when stale or when the review count has changed. Returns
// nil (no error) if the product has no reviews or if AI is disabled and no
// cached summary exists.
func (s *ProductAIService) ReviewSummary(productID uuid.UUID) (*models.ProductReviewSummary, error) {
	// Count live reviews first.
	var liveCount int64
	if err := s.db.Model(&models.Review{}).
		Where("product_id = ?", productID).
		Count(&liveCount).Error; err != nil {
		return nil, fmt.Errorf("product_ai: count reviews: %w", err)
	}

	// Load the cached summary, if any.
	var cached models.ProductReviewSummary
	cacheErr := s.db.Where("product_id = ?", productID).First(&cached).Error
	hasCache := cacheErr == nil

	// No reviews? No summary.
	if liveCount == 0 {
		return nil, nil
	}

	// Return the cache if it's still fresh.
	if hasCache && int64(cached.ReviewCount) == liveCount &&
		time.Since(cached.GeneratedAt) < reviewSummaryMaxAge {
		return &cached, nil
	}

	// Need to (re)generate — but we can't without AI. Serve a stale cache if any.
	if !s.ai.IsEnabled() {
		if hasCache {
			return &cached, nil
		}
		return nil, nil
	}

	reviews, _, err := s.reviewRepo.GetByProduct(productID, 1, 50) // up to 50 most recent
	if err != nil {
		return nil, fmt.Errorf("product_ai: load reviews: %w", err)
	}
	if len(reviews) == 0 {
		return nil, nil
	}

	prompt := buildReviewSummaryPrompt(reviews)
	summary, err := s.ai.callGemini(prompt, 180)
	if err != nil {
		if hasCache {
			return &cached, nil
		}
		return nil, err
	}

	record := models.ProductReviewSummary{
		ProductID:   productID,
		Summary:     strings.TrimSpace(summary),
		ReviewCount: int(liveCount),
		GeneratedAt: time.Now(),
	}
	// Upsert.
	if err := s.db.Save(&record).Error; err != nil {
		return &record, nil // ok to surface the summary even if persistence failed
	}
	return &record, nil
}

func buildReviewSummaryPrompt(reviews []models.Review) string {
	var b strings.Builder
	b.WriteString(`You summarise product reviews for an online art marketplace. Based ONLY on the reviews below, write ONE concise paragraph (2-3 sentences, under 280 characters) describing the consistent praise and concerns buyers raise. Use a neutral, factual tone — no marketing language. Do not invent anything not in the reviews. If reviews broadly agree, say so. If they conflict, note the disagreement. Output just the paragraph text, no preamble.

Reviews:
`)
	for i, r := range reviews {
		comment := ""
		if r.Comment != nil {
			comment = *r.Comment
		}
		title := ""
		if r.Title != nil {
			title = *r.Title
		}
		fmt.Fprintf(&b, "\n[%d] %d/5", i+1, r.Rating)
		if title != "" {
			fmt.Fprintf(&b, " — %s", title)
		}
		if comment != "" {
			fmt.Fprintf(&b, "\n%s", comment)
		}
		b.WriteString("\n")
	}
	return b.String()
}
