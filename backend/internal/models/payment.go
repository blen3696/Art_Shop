package models

import (
	"time"

	"github.com/google/uuid"
)

// Payment is a single attempt to charge a buyer for an order through a payment
// provider (Chapa today; Stripe/Paystack later). An order may have several
// payment rows — one per attempt — and only the latest successful one counts.
type Payment struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrderID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TxRef         string     `gorm:"size:64;uniqueIndex;not null" json:"tx_ref"`
	ProviderRef   *string    `gorm:"size:128" json:"provider_ref"`
	Provider      string     `gorm:"size:30;not null;default:chapa" json:"provider"`
	Amount        float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	Currency      string     `gorm:"size:10;not null;default:ETB" json:"currency"`
	Status        string     `gorm:"size:20;not null;default:pending;index" json:"status"`
	CheckoutURL   *string    `gorm:"type:text" json:"checkout_url"`
	RawResponse   []byte     `gorm:"type:jsonb" json:"-"`
	FailureReason *string    `gorm:"type:text" json:"failure_reason"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

// TableName overrides the default table name.
func (Payment) TableName() string {
	return "payments"
}

// Terminal returns true when the payment is in a final state that should not
// be mutated further. Used to make webhook processing idempotent.
func (p *Payment) Terminal() bool {
	return p.Status == "success" || p.Status == "failed" || p.Status == "cancelled"
}

// Valid payment statuses.
const (
	PaymentStatusPending   = "pending"
	PaymentStatusSuccess   = "success"
	PaymentStatusFailed    = "failed"
	PaymentStatusCancelled = "cancelled"
)

// InitializePaymentRequest is the HTTP payload to start a checkout for an order.
type InitializePaymentRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
}

// InitializePaymentResponse is what we return to the frontend after creating a
// payment attempt. The frontend redirects the user to CheckoutURL.
type InitializePaymentResponse struct {
	PaymentID   uuid.UUID `json:"payment_id"`
	TxRef       string    `json:"tx_ref"`
	CheckoutURL string    `json:"checkout_url"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
}