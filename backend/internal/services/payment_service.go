package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentService orchestrates the payment flow: it owns the state machine
// for Payment rows and coordinates between the Chapa HTTP client, the order
// repo, and the payment repo. Handlers talk to this service — never to
// ChapaService directly — so the business rules (authorization, order
// already paid, amount mismatch) live in exactly one place.
type PaymentService struct {
	payments  *repository.PaymentRepository
	orders    *repository.OrderRepository
	users     *repository.UserRepository
	chapa     *ChapaService
	cfg       *config.Config
}

func NewPaymentService(
	payments *repository.PaymentRepository,
	orders *repository.OrderRepository,
	users *repository.UserRepository,
	chapa *ChapaService,
	cfg *config.Config,
) *PaymentService {
	return &PaymentService{
		payments: payments,
		orders:   orders,
		users:    users,
		chapa:    chapa,
		cfg:      cfg,
	}
}

// Initialize creates a new Payment row and a Chapa checkout session for the
// given order. The caller must be the buyer of the order — we verify that.
// Returns the checkout URL the frontend should redirect the user to.
func (s *PaymentService) Initialize(ctx context.Context, buyerID, orderID uuid.UUID) (*models.InitializePaymentResponse, error) {
	if !s.chapa.IsEnabled() {
		return nil, fmt.Errorf("payment_service: chapa is not configured")
	}

	order, err := s.orders.FindByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.BuyerID != buyerID {
		return nil, fmt.Errorf("not authorized to pay for this order")
	}
	if order.PaymentStatus == "paid" {
		return nil, fmt.Errorf("order is already paid")
	}
	if order.Status == "cancelled" {
		return nil, fmt.Errorf("cannot pay for a cancelled order")
	}
	if order.Total <= 0 {
		return nil, fmt.Errorf("order total is zero")
	}

	buyer, err := s.users.FindByID(buyerID)
	if err != nil || buyer == nil {
		return nil, fmt.Errorf("buyer not found")
	}

	// Create the Payment row FIRST — its UUID becomes tx_ref. If the Chapa
	// call later fails, we still have a pending row documenting the attempt,
	// which shows up in the audit trail.
	payment := &models.Payment{
		OrderID:  order.ID,
		UserID:   buyerID,
		Provider: "chapa",
		Amount:   order.Total,
		Currency: s.chapa.Currency(),
		Status:   models.PaymentStatusPending,
	}
	payment.TxRef = uuid.New().String()
	if err := s.payments.Create(payment); err != nil {
		return nil, err
	}

	firstName, lastName := splitName(buyer.FullName)

	callbackBase := strings.TrimRight(s.cfg.ChapaCallbackBaseURL, "/")
	frontendBase := strings.TrimRight(s.cfg.AppURL, "/")

	req := InitializeRequest{
		Amount:      strconv.FormatFloat(order.Total, 'f', 2, 64),
		Currency:    s.chapa.Currency(),
		Email:       buyer.Email,
		FirstName:   firstName,
		LastName:    lastName,
		TxRef:       payment.TxRef,
		CallbackURL: callbackBase + "/api/payments/webhook",
		ReturnURL:   fmt.Sprintf("%s/payment/callback?tx_ref=%s", frontendBase, payment.TxRef),
	}
	req.Customization.Title = "ArtShop"
	req.Customization.Description = "Order " + order.OrderNumber

	// Shipping phone on the order is a better contact than the user profile
	// phone for this transaction.
	if order.ShippingPhone != nil {
		req.PhoneNumber = *order.ShippingPhone
	}

	result, err := s.chapa.Initialize(ctx, req)
	if err != nil {
		// Initialize failed before the user ever reached Chapa. Roll back the
		// order: restore stock and cancel the order, otherwise every failed
		// attempt permanently burns inventory. Best-effort — log and continue
		// if the rollback itself errors, the primary error still surfaces.
		if markErr := s.payments.MarkFailed(payment.ID, "chapa initialize failed: "+err.Error(), nil); markErr != nil {
			slog.Warn("payment_service: mark failed on init error", "payment_id", payment.ID, "error", markErr)
		}
		if cancelErr := s.orders.CancelAndRestoreStock(order.ID); cancelErr != nil {
			slog.Error("payment_service: cancel+restore stock failed", "order_id", order.ID, "error", cancelErr)
		}
		return nil, err
	}

	// Persist the checkout URL for audit / debugging. Not critical — if this
	// update fails we still return success so the user can still pay; we just
	// lose the cached URL.
	if err := s.payments.UpdateCheckoutURL(payment.ID, result.CheckoutURL, result.RawBody); err != nil {
		slog.Warn("payment_service: persist checkout url failed", "payment_id", payment.ID, "error", err)
	}

	return &models.InitializePaymentResponse{
		PaymentID:   payment.ID,
		TxRef:       payment.TxRef,
		CheckoutURL: result.CheckoutURL,
		Amount:      order.Total,
		Currency:    s.chapa.Currency(),
	}, nil
}

// HandleWebhook is called by the HTTP handler after it has already verified
// the HMAC signature. The webhook body is opaque to us — we trust ONLY the
// tx_ref inside it, then re-verify with Chapa. Never update DB state from
// the webhook body directly.
func (s *PaymentService) HandleWebhook(ctx context.Context, txRef string) error {
	if !s.chapa.IsEnabled() {
		return fmt.Errorf("payment_service: chapa is not configured")
	}
	if txRef == "" {
		return fmt.Errorf("payment_service: empty tx_ref in webhook")
	}
	return s.reconcile(ctx, txRef)
}

// VerifyByTxRef is called by the frontend after Chapa redirects the user
// back (e.g. /payment/callback?tx_ref=...). Safe to call many times — the
// repo update is idempotent.
func (s *PaymentService) VerifyByTxRef(ctx context.Context, txRef string) (*models.Payment, error) {
	if err := s.reconcile(ctx, txRef); err != nil {
		return nil, err
	}
	return s.payments.FindByTxRef(txRef)
}

// reconcile is the shared code path: fetch from Chapa's verify endpoint,
// compare amounts, then update DB. Called from both webhook and
// frontend-triggered verify paths.
func (s *PaymentService) reconcile(ctx context.Context, txRef string) error {
	payment, err := s.payments.FindByTxRef(txRef)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unknown tx_ref")
		}
		return err
	}

	// If the payment is already in a terminal state, reconcile is a no-op.
	// This is the idempotency shortcut for repeat webhooks.
	if payment.Terminal() {
		slog.Info("payment_service: reconcile skipped, already terminal",
			"tx_ref", txRef, "status", payment.Status)
		return nil
	}

	verify, err := s.chapa.Verify(ctx, txRef)
	if err != nil {
		return err
	}

	switch verify.Status {
	case "success":
		// Defense against amount tampering — even if Chapa's verify says
		// "success," reject if the amount doesn't match what we charged.
		if !amountsMatch(verify.Amount, payment.Amount) {
			reason := fmt.Sprintf("amount mismatch: expected %.2f, got %.2f", payment.Amount, verify.Amount)
			slog.Error("payment_service: amount mismatch", "tx_ref", txRef, "expected", payment.Amount, "got", verify.Amount)
			if err := s.payments.MarkFailed(payment.ID, reason, verify.RawBody); err != nil {
				return err
			}
			if err := s.orders.CancelAndRestoreStock(payment.OrderID); err != nil {
				slog.Error("payment_service: cancel+restore stock after amount mismatch",
					"order_id", payment.OrderID, "error", err)
			}
			return nil
		}
		if err := s.payments.MarkSuccess(payment.ID, verify.ProviderRef, verify.RawBody); err != nil {
			return err
		}
		// Cascade to the order: mark paid + confirmed, and stash the provider ref
		// on the order for easy lookup from the orders table.
		if err := s.orders.MarkPaid(payment.OrderID, verify.ProviderRef); err != nil {
			slog.Error("payment_service: order mark paid failed", "order_id", payment.OrderID, "error", err)
			return err
		}
		slog.Info("payment_service: payment succeeded", "tx_ref", txRef, "order_id", payment.OrderID)
		return nil

	case "failed":
		if err := s.payments.MarkFailed(payment.ID, "chapa returned failed", verify.RawBody); err != nil {
			return err
		}
		// Release the stock back so the buyer (or another buyer) can retry.
		if err := s.orders.CancelAndRestoreStock(payment.OrderID); err != nil {
			slog.Error("payment_service: cancel+restore stock after verify failed",
				"order_id", payment.OrderID, "error", err)
		}
		return nil

	default:
		// "pending" or unknown — don't mutate. The next webhook/verify will
		// retry. Logging "unknown" keeps us alerted if Chapa adds new states.
		slog.Info("payment_service: verify still pending", "tx_ref", txRef, "status", verify.Status)
		return nil
	}
}

// amountsMatch compares two decimal amounts within a cent of tolerance —
// float rounding off the wire shouldn't cause spurious mismatches.
func amountsMatch(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.01
}

// splitName takes a full name and returns (first, last). Chapa requires both.
// A single-word name becomes (name, "-") because Chapa rejects empty last_name.
func splitName(full string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(full))
	switch len(parts) {
	case 0:
		return "Customer", "-"
	case 1:
		return parts[0], "-"
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}
