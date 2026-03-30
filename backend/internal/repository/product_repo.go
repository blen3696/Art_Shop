package repository

import (
	"fmt"

	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductQueryParams holds all possible filters for product listing queries.
type ProductQueryParams struct {
	Page       int
	PerPage    int
	CategoryID string
	MinPrice   float64
	MaxPrice   float64
	Medium     string
	SortBy     string
	SortOrder  string
	SellerID   string
	Search     string
	Featured   bool
}

// ProductRepository handles all database operations for products and categories.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository instance.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create inserts a new product record into the database.
func (r *ProductRepository) Create(product *models.Product) error {
	if err := r.db.Create(product).Error; err != nil {
		return fmt.Errorf("product_repo: failed to create product: %w", err)
	}
	return nil
}

// FindByID retrieves a product by its UUID, preloading Seller and Category.
func (r *ProductRepository) FindByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	if err := r.db.Preload("Seller").Preload("Category").First(&product, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("product_repo: product not found: %w", err)
	}
	return &product, nil
}

// Update saves changes to an existing product record.
func (r *ProductRepository) Update(product *models.Product) error {
	if err := r.db.Save(product).Error; err != nil {
		return fmt.Errorf("product_repo: failed to update product: %w", err)
	}
	return nil
}

// Delete removes a product by its UUID.
func (r *ProductRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.Product{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("product_repo: failed to delete product: %w", err)
	}
	return nil
}

// List returns a paginated, filtered list of products based on the provided query parameters.
func (r *ProductRepository) List(params ProductQueryParams) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).Where("is_published = ?", true)

	if params.CategoryID != "" {
		catID, err := uuid.Parse(params.CategoryID)
		if err == nil {
			query = query.Where("category_id = ?", catID)
		}
	}

	if params.MinPrice > 0 {
		query = query.Where("price >= ?", params.MinPrice)
	}

	if params.MaxPrice > 0 {
		query = query.Where("price <= ?", params.MaxPrice)
	}

	if params.Medium != "" {
		query = query.Where("medium = ?", params.Medium)
	}

	if params.SellerID != "" {
		sellerID, err := uuid.Parse(params.SellerID)
		if err == nil {
			query = query.Where("seller_id = ?", sellerID)
		}
	}

	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if params.Featured {
		query = query.Where("is_featured = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("product_repo: failed to count products: %w", err)
	}

	// Determine sort order.
	sortBy := "created_at"
	sortOrder := "DESC"
	allowedSorts := map[string]bool{
		"created_at": true, "price": true, "title": true,
		"avg_rating": true, "total_sales": true, "view_count": true,
	}
	if params.SortBy != "" && allowedSorts[params.SortBy] {
		sortBy = params.SortBy
	}
	if params.SortOrder == "asc" || params.SortOrder == "ASC" {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage
	if err := query.
		Preload("Seller").
		Preload("Category").
		Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Offset(offset).
		Limit(params.PerPage).
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("product_repo: failed to list products: %w", err)
	}

	return products, total, nil
}

// ListBySeller returns a paginated list of products for a specific seller.
func (r *ProductRepository) ListBySeller(sellerID uuid.UUID, page, perPage int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).Where("seller_id = ?", sellerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("product_repo: failed to count seller products: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.
		Preload("Category").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("product_repo: failed to list seller products: %w", err)
	}

	return products, total, nil
}

// Search performs a full-text search on product titles and descriptions using PostgreSQL tsvector.
func (r *ProductRepository) Search(query string, page, perPage int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	searchQuery := r.db.Model(&models.Product{}).
		Where("is_published = ?", true).
		Where(
			"to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?) OR title ILIKE ?",
			query, "%"+query+"%",
		)

	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("product_repo: search count failed: %w", err)
	}

	offset := (page - 1) * perPage
	orderClause := gorm.Expr(
		"ts_rank(to_tsvector('english', title || ' ' || COALESCE(description, '')), plainto_tsquery('english', ?)) DESC",
		query,
	)
	if err := searchQuery.
		Preload("Seller").
		Preload("Category").
		Order(orderClause).
		Offset(offset).
		Limit(perPage).
		Find(&products).Error; err != nil {
		// Fall back to ILIKE if ts_rank ordering fails.
		searchQuery2 := r.db.Model(&models.Product{}).
			Where("is_published = ?", true).
			Where("title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")

		if err2 := searchQuery2.
			Preload("Seller").
			Preload("Category").
			Order("created_at DESC").
			Offset(offset).
			Limit(perPage).
			Find(&products).Error; err2 != nil {
			return nil, 0, fmt.Errorf("product_repo: search failed: %w", err2)
		}
	}

	return products, total, nil
}

// GetFeatured returns up to `limit` featured products.
func (r *ProductRepository) GetFeatured(limit int) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.
		Where("is_featured = ? AND is_published = ?", true, true).
		Preload("Seller").
		Preload("Category").
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("product_repo: failed to get featured products: %w", err)
	}
	return products, nil
}

// IncrementViewCount atomically increments the view count for a product.
func (r *ProductRepository) IncrementViewCount(id uuid.UUID) error {
	if err := r.db.Model(&models.Product{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		return fmt.Errorf("product_repo: failed to increment view count: %w", err)
	}
	return nil
}

// GetCategories returns all active categories ordered by sort_order.
func (r *ProductRepository) GetCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("product_repo: failed to get categories: %w", err)
	}
	return categories, nil
}
