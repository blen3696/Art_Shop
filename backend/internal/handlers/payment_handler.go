package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/services"
	"github.com/artshop/backend/pkg/response"
)

// PaymentHandler exposes HTTP endpoints for starting and reconciling payments.
//
// Routes:
//   POST /api/payments/initialize   — auth required, body {order_id}
//   POST /api/payments/webhook      — PUBLIC, secured by HMAC signature
//   GET  /api/payments/verify       — auth required, query ?tx_ref=...
type PaymentHandler struct {
	payments *services.PaymentService
	chapa    *services.ChapaService
}

func NewPaymentHandler(payments *services.PaymentService, chapa *services.ChapaService) *PaymentHandler {
	return &PaymentHandler{payments: payments, chapa: chapa}
}

// Initialize handles POST /api/payments/initialize — creates a Chapa checkout
// session for the given order. Returns the hosted checkout URL.
func (h *PaymentHandler) Initialize(w http.ResponseWriter, r *http.Request) {
	if !h.chapa.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "PAYMENTS_DISABLED",
			"Chapa payments are not configured")
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	var req models.InitializePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	if req.OrderID.String() == "00000000-0000-0000-0000-000000000000" {
		response.ValidationError(w, map[string]string{"order_id": "order_id is required"})
		return
	}

	result, err := h.payments.Initialize(r.Context(), userID, req.OrderID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "PAYMENT_INIT_FAILED", err.Error())
		return
	}

	response.Created(w, result)
}

// Webhook handles POST /api/payments/webhook — called by Chapa server-to-server
// when a payment changes state. The body is HMAC-SHA256 signed with the
// webhook secret we configured in the Chapa dashboard. We:
//   1. Read the raw body (must stay exactly as Chapa sent it for the HMAC).
//   2. Verify the signature.
//   3. Extract tx_ref.
//   4. Hand off to the service, which calls Chapa's verify endpoint for truth.
//
// We ALWAYS return 200 once the signature is valid — even if reconciliation
// fails internally — because Chapa retries on non-2xx and we don't want
// retry storms for transient DB hiccups. Internal errors are logged loudly.
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if !h.chapa.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "PAYMENTS_DISABLED",
			"Chapa payments are not configured")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "cannot read body")
		return
	}
	defer r.Body.Close()

	// Chapa uses either "Chapa-Signature" or "x-chapa-signature" depending on
	// dashboard version. Accept both.
	sig := r.Header.Get("Chapa-Signature")
	if sig == "" {
		sig = r.Header.Get("x-chapa-signature")
	}
	if !h.chapa.VerifyWebhookSignature(body, sig) {
		slog.Warn("payment_handler: rejected webhook with bad signature",
			"ip", r.RemoteAddr, "sig_present", sig != "")
		response.Error(w, http.StatusUnauthorized, "BAD_SIGNATURE", "invalid webhook signature")
		return
	}

	// Minimal parse — we only trust tx_ref from the body; everything else
	// comes from Chapa's verify endpoint.
	var payload struct {
		TxRef string `json:"tx_ref"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Warn("payment_handler: webhook JSON parse failed", "error", err)
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "cannot parse webhook body")
		return
	}

	if err := h.payments.HandleWebhook(r.Context(), payload.TxRef); err != nil {
		slog.Error("payment_handler: webhook reconcile failed",
			"tx_ref", payload.TxRef, "error", err)
		// Still 200 — Chapa doesn't need to retry; we'll reconcile via verify.
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// Verify handles GET /api/payments/verify?tx_ref=... — called by the frontend
// after Chapa redirects the user back. Triggers reconciliation and returns
// the current payment state so the UI can show success/failure.
func (h *PaymentHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if !h.chapa.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "PAYMENTS_DISABLED",
			"Chapa payments are not configured")
		return
	}

	txRef := r.URL.Query().Get("tx_ref")
	if txRef == "" {
		response.ValidationError(w, map[string]string{"tx_ref": "tx_ref is required"})
		return
	}

	payment, err := h.payments.VerifyByTxRef(r.Context(), txRef)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VERIFY_FAILED", err.Error())
		return
	}

	// Defense-in-depth: a logged-in user should only see their own payments.
	userID := middleware.GetUserIDFromContext(r.Context())
	if payment.UserID != userID {
		response.Error(w, http.StatusForbidden, "FORBIDDEN", "not your payment")
		return
	}

	response.JSON(w, http.StatusOK, payment)
}
