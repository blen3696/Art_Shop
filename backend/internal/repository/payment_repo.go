package repository

import (
	"fmt"
	"time"

	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentRepository handles all database operations for the payments table.
type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create inserts a new payment row. Caller is expected to have already set
// Amount, Currency, OrderID, UserID, TxRef. Status defaults to "pending".
func (r *PaymentRepository) Create(p *models.Payment) error {
	if err := r.db.Create(p).Error; err != nil {
		return fmt.Errorf("payment_repo: create: %w", err)
	}
	return nil
}

// FindByTxRef looks up a payment by its tx_ref (our UUID sent to Chapa).
// Returns gorm.ErrRecordNotFound when not found — callers should check that
// explicitly if they want a 404 vs 500.
func (r *PaymentRepository) FindByTxRef(txRef string) (*models.Payment, error) {
	var p models.Payment
	if err := r.db.Where("tx_ref = ?", txRef).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// FindByID looks up a payment by its primary key.
func (r *PaymentRepository) FindByID(id uuid.UUID) (*models.Payment, error) {
	var p models.Payment
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// MarkSuccess transitions a payment to success atomically, using an
// optimistic guard (status = 'pending') so a duplicate webhook won't
// re-overwrite a completed row. RowsAffected == 0 means the payment was
// already terminal, which the caller should treat as success (idempotent).
//
// rawResponse is converted to string before writing: under Supabase's simple
// query protocol, pgx encodes []byte as a bytea literal which Postgres refuses
// to cast to jsonb. Text works because jsonb has an implicit cast from text.
func (r *PaymentRepository) MarkSuccess(id uuid.UUID, providerRef string, rawResponse []byte) error {
	now := time.Now()
	updates := map[string]any{
		"status":       "success",
		"provider_ref": providerRef,
		"completed_at": &now,
		"updated_at":   now,
	}
	if len(rawResponse) > 0 {
		updates["raw_response"] = string(rawResponse)
	}
	result := r.db.Model(&models.Payment{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("payment_repo: mark success: %w", result.Error)
	}
	return nil
}

// MarkFailed transitions a payment to failed. Same optimistic guard as success
// to keep the operation idempotent.
func (r *PaymentRepository) MarkFailed(id uuid.UUID, reason string, rawResponse []byte) error {
	now := time.Now()
	updates := map[string]any{
		"status":         "failed",
		"failure_reason": reason,
		"completed_at":   &now,
		"updated_at":     now,
	}
	if len(rawResponse) > 0 {
		updates["raw_response"] = string(rawResponse)
	}
	result := r.db.Model(&models.Payment{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("payment_repo: mark failed: %w", result.Error)
	}
	return nil
}

// UpdateCheckoutURL stores the Chapa-returned checkout URL and the raw
// initialize response on a pending payment. Called right after a successful
// Chapa initialize call — purely informational, not part of the critical path.
func (r *PaymentRepository) UpdateCheckoutURL(id uuid.UUID, checkoutURL string, rawResponse []byte) error {
	updates := map[string]any{
		"checkout_url": checkoutURL,
	}
	if len(rawResponse) > 0 {
		updates["raw_response"] = string(rawResponse)
	}
	result := r.db.Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("payment_repo: update checkout url: %w", result.Error)
	}
	return nil
}

// ListByOrder returns all payment attempts for an order, newest first.
// Used by the frontend to show "Previous attempt failed, try again" UX.
func (r *PaymentRepository) ListByOrder(orderID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.db.
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("payment_repo: list by order: %w", err)
	}
	return payments, nil
}
