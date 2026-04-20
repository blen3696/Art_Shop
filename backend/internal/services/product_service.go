package services

import (
	"fmt"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateProductRequest holds the data required to create a new product.
type CreateProductRequest struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Price          float64  `json:"price"`
	CompareAtPrice *float64 `json:"compare_at_price"`
	CategoryID     string   `json:"category_id"`
	Images         []string `json:"images"`
	Thumbnail      string   `json:"thumbnail"`
	Stock          int      `json:"stock"`
	SKU            string   `json:"sku"`
	Tags           []string `json:"tags"`
	Medium         string   `json:"medium"`
	Dimensions     string   `json:"dimensions"`
	Weight         *float64 `json:"weight"`
}

// UpdateProductRequest holds the data for updating an existing product.
type UpdateProductRequest struct {
	Title          *string  `json:"title"`
	Description    *string  `json:"description"`
	Price          *float64 `json:"price"`
	CompareAtPrice *float64 `json:"compare_at_price"`
	CategoryID     *string  `json:"category_id"`
	Images         []string `json:"images"`
	Thumbnail      *string  `json:"thumbnail"`
	Stock          *int     `json:"stock"`
	SKU            *string  `json:"sku"`
	Tags           []string `json:"tags"`
	Medium         *string  `json:"medium"`
	Dimensions     *string  `json:"dimensions"`
	Weight         *float64 `json:"weight"`
	IsPublished    *bool    `json:"is_published"`
}

// ProductService handles business logic for products.
type ProductService struct {
	productRepo *repository.ProductRepository
	productAI   *ProductAIService
	cfg         *config.Config
}

// NewProductService creates a new ProductService instance. productAI may be nil
// when AI is not configured — embedding calls no-op in that case.
func NewProductService(
	productRepo *repository.ProductRepository,
	productAI *ProductAIService,
	cfg *config.Config,
) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		productAI:   productAI,
		cfg:         cfg,
	}
}

// Create creates a new product for the given seller.
func (s *ProductService) Create(sellerID uuid.UUID, req CreateProductRequest) (*models.Product, error) {
	product := &models.Product{
		SellerID:    sellerID,
		Title:       req.Title,
		Price:       req.Price,
		Stock:       req.Stock,
		IsPublished: true,
	}

	if req.Description != "" {
		product.Description = &req.Description
	}
	if req.CompareAtPrice != nil {
		product.CompareAtPrice = req.CompareAtPrice
	}
	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err == nil {
			product.CategoryID = &catID
		}
	}
	if len(req.Images) > 0 {
		product.Images = pq.StringArray(req.Images)
	}
	if req.Thumbnail != "" {
		product.Thumbnail = &req.Thumbnail
	}
	if req.SKU != "" {
		product.SKU = &req.SKU
	}
	if len(req.Tags) > 0 {
		product.Tags = pq.StringArray(req.Tags)
	}
	if req.Medium != "" {
		product.Medium = &req.Medium
	}
	if req.Dimensions != "" {
		product.Dimensions = &req.Dimensions
	}
	if req.Weight != nil {
		product.Weight = req.Weight
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("product_service: failed to create product: %w", err)
	}

	// Generate the semantic embedding asynchronously so the response isn't
	// blocked on an external API call.
	if s.productAI != nil {
		s.productAI.EmbedProductAsync(product)
	}

	// Re-fetch to get preloaded associations.
	return s.productRepo.FindByID(product.ID)
}

// Update modifies an existing product, verifying the seller owns it.
func (s *ProductService) Update(sellerID uuid.UUID, productID uuid.UUID, req UpdateProductRequest) (*models.Product, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Verify ownership (admins can update any product via handlers).
	if product.SellerID != sellerID {
		return nil, fmt.Errorf("not authorized to update this product")
	}

	// Apply updates.
	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.CompareAtPrice != nil {
		product.CompareAtPrice = req.CompareAtPrice
	}
	if req.CategoryID != nil {
		if *req.CategoryID != "" {
			catID, err := uuid.Parse(*req.CategoryID)
			if err == nil {
				product.CategoryID = &catID
			}
		} else {
			product.CategoryID = nil
		}
	}
	if req.Images != nil {
		product.Images = pq.StringArray(req.Images)
	}
	if req.Thumbnail != nil {
		product.Thumbnail = req.Thumbnail
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.SKU != nil {
		product.SKU = req.SKU
	}
	if req.Tags != nil {
		product.Tags = pq.StringArray(req.Tags)
	}
	if req.Medium != nil {
		product.Medium = req.Medium
	}
	if req.Dimensions != nil {
		product.Dimensions = req.Dimensions
	}
	if req.Weight != nil {
		product.Weight = req.Weight
	}
	if req.IsPublished != nil {
		product.IsPublished = *req.IsPublished
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("product_service: failed to update product: %w", err)
	}

	// Re-embed on update in case title/description/tags/medium changed.
	if s.productAI != nil {
		s.productAI.EmbedProductAsync(product)
	}

	return s.productRepo.FindByID(product.ID)
}

// Delete removes a product, verifying the seller owns it.
func (s *ProductService) Delete(sellerID uuid.UUID, productID uuid.UUID) error {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	if product.SellerID != sellerID {
		return fmt.Errorf("not authorized to delete this product")
	}

	return s.productRepo.Delete(productID)
}

// GetByID retrieves a single product by ID and increments its view count.
func (s *ProductService) GetByID(id uuid.UUID) (*models.Product, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Increment view count asynchronously (best effort).
	go func() {
		_ = s.productRepo.IncrementViewCount(id)
	}()

	return product, nil
}

// List returns a filtered, paginated list of products.
func (s *ProductService) List(params repository.ProductQueryParams) ([]models.Product, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}
	return s.productRepo.List(params)
}

// Search returns products matching a query. Uses semantic (vector) search when
// AI is configured, otherwise falls back to PostgreSQL full-text search.
func (s *ProductService) Search(query string, page, perPage int) ([]models.Product, int64, error) {
	if s.productAI != nil {
		return s.productAI.SearchWithFallback(query, page, perPage)
	}
	return s.searchKeyword(query, page, perPage)
}

// searchKeyword is the original full-text search path.
func (s *ProductService) searchKeyword(query string, page, perPage int) ([]models.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.productRepo.Search(query, page, perPage)
}

// GetFeatured returns featured products.
func (s *ProductService) GetFeatured(limit int) ([]models.Product, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.productRepo.GetFeatured(limit)
}

// GetCategories returns all active product categories.
func (s *ProductService) GetCategories() ([]models.Category, error) {
	return s.productRepo.GetCategories()
}
