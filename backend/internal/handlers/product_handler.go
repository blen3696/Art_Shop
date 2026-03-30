package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/repository"
	"github.com/artshop/backend/internal/services"
	"github.com/artshop/backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ProductHandler handles HTTP requests for product endpoints.
type ProductHandler struct {
	productService *services.ProductService
}

// NewProductHandler creates a new ProductHandler instance.
func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// List handles GET /api/products with query parameter filters.
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	minPrice, _ := strconv.ParseFloat(q.Get("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(q.Get("max_price"), 64)

	params := repository.ProductQueryParams{
		Page:       page,
		PerPage:    perPage,
		CategoryID: q.Get("category"),
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		Medium:     q.Get("medium"),
		SortBy:     q.Get("sort_by"),
		SortOrder:  q.Get("sort_order"),
		SellerID:   q.Get("seller_id"),
		Search:     q.Get("search"),
		Featured:   q.Get("featured") == "true",
	}

	products, total, err := h.productService.List(params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	response.Paginated(w, http.StatusOK, products, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/products/:id.
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	product, err := h.productService.GetByID(id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Product not found")
		return
	}

	response.JSON(w, http.StatusOK, product)
}

// Create handles POST /api/products (requires seller/admin).
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var req services.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields.
	errors := make(map[string]string)
	if req.Title == "" {
		errors["title"] = "Title is required"
	}
	if req.Price <= 0 {
		errors["price"] = "Price must be greater than 0"
	}
	if len(errors) > 0 {
		response.ValidationError(w, errors)
		return
	}

	product, err := h.productService.Create(userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	response.Created(w, product)
}

// Update handles PUT /api/products/:id (requires seller/admin, verifies ownership).
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	role := middleware.GetUserRoleFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req services.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Admins can update any product by fetching the real seller ID first.
	sellerID := userID
	if role == "admin" {
		existing, err := h.productService.GetByID(productID)
		if err != nil {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		sellerID = existing.SellerID
	}

	product, err := h.productService.Update(sellerID, productID, req)
	if err != nil {
		if err.Error() == "not authorized to update this product" {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		if err.Error() == "product not found" {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, product)
}

// Delete handles DELETE /api/products/:id (requires seller/admin).
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	role := middleware.GetUserRoleFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	// Admins can delete any product by fetching the real seller ID first.
	sellerID := userID
	if role == "admin" {
		existing, err := h.productService.GetByID(productID)
		if err != nil {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		sellerID = existing.SellerID
	}

	if err := h.productService.Delete(sellerID, productID); err != nil {
		if err.Error() == "not authorized to delete this product" {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	response.NoContent(w)
}

// Featured handles GET /api/products/featured.
func (h *ProductHandler) Featured(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	products, err := h.productService.GetFeatured(limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, products)
}

// Search handles GET /api/products/search?q=.
func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := q.Get("q")
	if query == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Search query 'q' is required")
		return
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	products, total, err := h.productService.Search(query, page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "SEARCH_FAILED", err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	response.Paginated(w, http.StatusOK, products, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

// Categories handles GET /api/categories.
func (h *ProductHandler) Categories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.productService.GetCategories()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, categories)
}
