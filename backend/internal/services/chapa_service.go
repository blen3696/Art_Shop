package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/artshop/backend/internal/config"
)

// chapaBaseURL is Chapa's production API host. Test keys (CHASECK_TEST_...) and
// live keys (CHASECK_...) both hit the same host — the key prefix determines
// which environment you're in, not the URL.
const chapaBaseURL = "https://api.chapa.co/v1"

// ChapaService is a thin client over Chapa's REST API. It exposes only the
// two endpoints we use: Initialize (creates a hosted checkout session) and
// Verify (looks up a transaction by tx_ref). We talk to Chapa directly with
// net/http rather than pulling in a third-party SDK — fewer deps, and
// Chapa's API surface is small enough that a hand-rolled client stays clear.
//
// If CHAPA_SECRET_KEY is empty the service is disabled: IsEnabled() returns
// false and every method returns an error. Callers check IsEnabled() before
// invoking, exactly like EmailService.
type ChapaService struct {
	secretKey     string
	webhookSecret string
	currency      string
	enabled       bool
	http          *http.Client
}

// NewChapaService constructs the client from config. Missing secret key
// disables the service rather than failing hard — staying consistent with the
// rest of the codebase (AI, email both degrade gracefully).
func NewChapaService(cfg *config.Config) *ChapaService {
	if cfg.ChapaSecretKey == "" {
		slog.Warn("chapa_service: CHAPA_SECRET_KEY not set — Chapa checkout disabled")
		return &ChapaService{enabled: false}
	}
	return &ChapaService{
		secretKey:     cfg.ChapaSecretKey,
		webhookSecret: cfg.ChapaWebhookSecret,
		currency:      cfg.ChapaCurrency,
		enabled:       true,
		http:          &http.Client{Timeout: 15 * time.Second},
	}
}

// IsEnabled reports whether Chapa is configured. Handlers use this to decide
// whether to 503 the endpoint instead of returning a cryptic upstream error.
func (s *ChapaService) IsEnabled() bool { return s.enabled }

// Currency returns the configured currency code (ETB by default). Exposed so
// the payment service can stamp it onto Payment rows.
func (s *ChapaService) Currency() string { return s.currency }

// ----------------------------------------------------------------------------
// Initialize — creates a hosted checkout session.
// ----------------------------------------------------------------------------

// InitializeRequest is the payload Chapa's /transaction/initialize accepts.
// Field names match Chapa's JSON exactly — don't rename without updating tags.
type InitializeRequest struct {
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	TxRef       string `json:"tx_ref"`
	CallbackURL string `json:"callback_url"` // server-to-server webhook
	ReturnURL   string `json:"return_url"`   // browser redirect after payment
	Customization struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"customization,omitempty"`
}

// initializeResponse mirrors Chapa's JSON response for initialize calls.
// We keep this private — callers only need CheckoutURL from InitializeResult.
type initializeResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}

// InitializeResult is what the service returns to the caller.
type InitializeResult struct {
	CheckoutURL string
	RawBody     []byte // the full response, stored in payments.raw_response
}

// Initialize asks Chapa to create a hosted checkout session and returns the
// URL we redirect the buyer to. Amount is passed as a string because Chapa's
// API validates it as a string; 2 decimal places is safe for any fiat.
func (s *ChapaService) Initialize(ctx context.Context, req InitializeRequest) (*InitializeResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("chapa_service: disabled (no secret key)")
	}
	if req.Currency == "" {
		req.Currency = s.currency
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("chapa_service: marshal initialize: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		chapaBaseURL+"/transaction/initialize", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("chapa_service: build initialize request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chapa_service: initialize HTTP: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("chapa_service: initialize rejected",
			"status", resp.StatusCode, "body", string(raw), "tx_ref", req.TxRef)
		return nil, fmt.Errorf("chapa_service: chapa returned %d: %s", resp.StatusCode, string(raw))
	}

	var parsed initializeResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("chapa_service: parse initialize: %w", err)
	}
	if parsed.Status != "success" || parsed.Data.CheckoutURL == "" {
		return nil, fmt.Errorf("chapa_service: initialize not successful: %s", parsed.Message)
	}

	slog.Info("chapa_service: checkout session created",
		"tx_ref", req.TxRef, "amount", req.Amount, "currency", req.Currency)

	return &InitializeResult{
		CheckoutURL: parsed.Data.CheckoutURL,
		RawBody:     raw,
	}, nil
}

// ----------------------------------------------------------------------------
// Verify — server-side source of truth for transaction status.
// ----------------------------------------------------------------------------

// VerifyResult is the parsed, trust-me-now view of a transaction. We use the
// amount Chapa reports (not the amount the caller claims) to defend against
// tampering.
type VerifyResult struct {
	Status      string  // "success" | "failed" | "pending"
	TxRef       string
	ProviderRef string  // Chapa's own reference
	Amount      float64
	Currency    string
	RawBody     []byte
}

type verifyResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    struct {
		Status    string `json:"status"`
		TxRef     string `json:"tx_ref"`
		Reference string `json:"reference"`
		Amount    any    `json:"amount"`   // Chapa sometimes sends string, sometimes number
		Currency  string `json:"currency"`
	} `json:"data"`
}

// Verify calls GET /transaction/verify/:tx_ref and returns the canonical
// transaction status. Always prefer this over trusting webhook bodies — the
// webhook signals "something happened," Verify tells you what.
func (s *ChapaService) Verify(ctx context.Context, txRef string) (*VerifyResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("chapa_service: disabled (no secret key)")
	}
	if txRef == "" {
		return nil, fmt.Errorf("chapa_service: tx_ref is required")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		chapaBaseURL+"/transaction/verify/"+txRef, nil)
	if err != nil {
		return nil, fmt.Errorf("chapa_service: build verify request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.secretKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chapa_service: verify HTTP: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("chapa_service: verify non-2xx",
			"status", resp.StatusCode, "tx_ref", txRef, "body", string(raw))
		return nil, fmt.Errorf("chapa_service: chapa verify returned %d: %s", resp.StatusCode, string(raw))
	}

	var parsed verifyResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("chapa_service: parse verify: %w", err)
	}

	amount, err := toFloat(parsed.Data.Amount)
	if err != nil {
		return nil, fmt.Errorf("chapa_service: parse amount: %w", err)
	}

	return &VerifyResult{
		Status:      parsed.Data.Status,
		TxRef:       parsed.Data.TxRef,
		ProviderRef: parsed.Data.Reference,
		Amount:      amount,
		Currency:    parsed.Data.Currency,
		RawBody:     raw,
	}, nil
}

// ----------------------------------------------------------------------------
// VerifyWebhookSignature — prevents attackers forging payment callbacks.
// ----------------------------------------------------------------------------

// VerifyWebhookSignature returns true iff HMAC-SHA256(body, webhookSecret)
// equals the hex-encoded signature the Chapa dashboard put in the header.
// Uses hmac.Equal for constant-time comparison (defeats timing attacks).
//
// If no webhook secret is configured the function returns false — callers
// should refuse the webhook in that case rather than falling back to "trust."
func (s *ChapaService) VerifyWebhookSignature(body []byte, providedSignature string) bool {
	if s.webhookSecret == "" || providedSignature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(providedSignature))
}

// toFloat converts Chapa's amount field (sometimes string, sometimes number)
// to a float64. We tolerate both because Chapa's schema isn't perfectly
// consistent across endpoints.
func toFloat(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case string:
		return strconv.ParseFloat(x, 64)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unexpected amount type %T", v)
	}
}
