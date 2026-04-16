package services

import (
	"fmt"
	"time"

	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DashboardStats holds aggregate statistics for the admin dashboard.
type DashboardStats struct {
	TotalUsers     int64           `json:"total_users"`
	TotalSellers   int64           `json:"total_sellers"`
	TotalProducts  int64           `json:"total_products"`
	TotalOrders    int64           `json:"total_orders"`
	TotalRevenue   float64         `json:"total_revenue"`
	RecentOrders   []models.Order  `json:"recent_orders"`
	MonthlyRevenue []MonthlyRevenue `json:"monthly_revenue"`
}

// MonthlyRevenue holds revenue data for a single month.
type MonthlyRevenue struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
}

// AdminService handles business logic for admin-specific operations.
type AdminService struct {
	userRepo    *repository.UserRepository
	productRepo *repository.ProductRepository
	orderRepo   *repository.OrderRepository
	db          *gorm.DB
}

// NewAdminService creates a new AdminService instance.
func NewAdminService(
	userRepo *repository.UserRepository,
	productRepo *repository.ProductRepository,
	orderRepo *repository.OrderRepository,
	db *gorm.DB,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		productRepo: productRepo,
		orderRepo:   orderRepo,
		db:          db,
	}
}

// GetDashboardStats returns comprehensive statistics for the admin dashboard.
func (s *AdminService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	if err := s.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("admin_service: failed to count users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("role = ?", "seller").Count(&stats.TotalSellers).Error; err != nil {
		return nil, fmt.Errorf("admin_service: failed to count sellers: %w", err)
	}

	if err := s.db.Model(&models.Product{}).Count(&stats.TotalProducts).Error; err != nil {
		return nil, fmt.Errorf("admin_service: failed to count products: %w", err)
	}

	if err := s.db.Model(&models.Order{}).Count(&stats.TotalOrders).Error; err != nil {
		return nil, fmt.Errorf("admin_service: failed to count orders: %w", err)
	}

	if err := s.db.Model(&models.Order{}).
		Select("COALESCE(SUM(total), 0)").
		Where("payment_status = ?", "paid").
		Row().Scan(&stats.TotalRevenue); err != nil {
		return nil, fmt.Errorf("admin_service: failed to get total revenue: %w", err)
	}

	if err := s.db.
		Preload("Items").
		Preload("Buyer").
		Order("created_at DESC").
		Limit(10).
		Find(&stats.RecentOrders).Error; err != nil {
		return nil, fmt.Errorf("admin_service: failed to get recent orders: %w", err)
	}

	// Monthly revenue for the last 12 months.
	monthlyRevenue, err := s.GetRevenueByMonth()
	if err != nil {
		return nil, err
	}
	stats.MonthlyRevenue = monthlyRevenue

	return stats, nil
}

// ListUsers returns a paginated list of users, optionally filtered by role.
func (s *AdminService) ListUsers(page, perPage int, role string) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.userRepo.List(page, perPage, role)
}

// UpdateUserRole changes a user's role.
func (s *AdminService) UpdateUserRole(userID uuid.UUID, role string) error {
	validRoles := map[string]bool{"buyer": true, "seller": true, "admin": true}
	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	user.Role = role
	return s.userRepo.Update(user)
}

// ToggleUserActive flips a user's is_active status.
func (s *AdminService) ToggleUserActive(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	user.IsActive = !user.IsActive
	return s.userRepo.Update(user)
}

// ListAllOrders returns a paginated list of all orders, optionally filtered by status.
func (s *AdminService) ListAllOrders(page, perPage int, status string) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.orderRepo.ListAll(page, perPage, status)
}

// ToggleProductFeatured flips a product's is_featured status.
func (s *AdminService) ToggleProductFeatured(productID uuid.UUID) error {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	product.IsFeatured = !product.IsFeatured
	return s.productRepo.Update(product)
}

// GetRevenueByMonth returns monthly revenue for the last 12 months.
func (s *AdminService) GetRevenueByMonth() ([]MonthlyRevenue, error) {
	type result struct {
		Month   string  `json:"month"`
		Revenue float64 `json:"revenue"`
	}

	var results []result

	// Query revenue grouped by month for the last 12 months.
	twelveMonthsAgo := time.Now().AddDate(-1, 0, 0)
	if err := s.db.Model(&models.Order{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') as month, COALESCE(SUM(total), 0) as revenue").
		Where("payment_status = ? AND created_at >= ?", "paid", twelveMonthsAgo).
		Group("TO_CHAR(created_at, 'YYYY-MM')").
		Order("month ASC").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("admin_service: failed to get monthly revenue: %w", err)
	}

	// Convert to MonthlyRevenue slice, filling in missing months with zero.
	revenueMap := make(map[string]float64)
	for _, r := range results {
		revenueMap[r.Month] = r.Revenue
	}

	monthly := make([]MonthlyRevenue, 0, 12)
	for i := 11; i >= 0; i-- {
		month := time.Now().AddDate(0, -i, 0).Format("2006-01")
		rev := revenueMap[month]
		monthly = append(monthly, MonthlyRevenue{
			Month:   month,
			Revenue: rev,
		})
	}

	return monthly, nil
}
