package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
)

// brevoEndpoint is Brevo's transactional email API.
const brevoEndpoint = "https://api.brevo.com/v3/smtp/email"

// EmailService sends transactional emails via Brevo's HTTP API. We talk to
// Brevo directly with net/http rather than pulling in the auto-generated
// SDK — fewer deps, smaller binary, easier to read.
//
// If BREVO_API_KEY is empty the service is disabled: calls log a warning and
// return nil so emails never block the request lifecycle.
type EmailService struct {
	apiKey    string
	fromName  string
	fromEmail string
	appURL    string
	enabled   bool
	http      *http.Client
}

func NewEmailService(cfg *config.Config) *EmailService {
	if cfg.BrevoAPIKey == "" {
		slog.Warn("email_service: BREVO_API_KEY not set — emails will be skipped")
		return &EmailService{enabled: false}
	}
	return &EmailService{
		apiKey:    cfg.BrevoAPIKey,
		fromName:  cfg.EmailFromName,
		fromEmail: cfg.EmailFromAddress,
		appURL:    strings.TrimRight(cfg.AppURL, "/"),
		enabled:   true,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

// SendWelcome emails a newly registered user.
func (s *EmailService) SendWelcome(ctx context.Context, to, name string) error {
	if !s.enabled || to == "" {
		return nil
	}
	return s.send(ctx, to, name, "Welcome to ArtShop", welcomeHTML(name))
}

// SendOrderConfirmation emails the buyer after a successful checkout.
func (s *EmailService) SendOrderConfirmation(ctx context.Context, to, name string, order *models.Order) error {
	if !s.enabled || to == "" {
		return nil
	}
	subject := fmt.Sprintf("Order confirmed — %s", order.OrderNumber)
	return s.send(ctx, to, name, subject, orderConfirmationHTML(name, order))
}

// SendOrderStatusUpdate emails the buyer when an order's status changes.
func (s *EmailService) SendOrderStatusUpdate(ctx context.Context, to, name string, order *models.Order) error {
	if !s.enabled || to == "" {
		return nil
	}
	subject := fmt.Sprintf("Order %s — %s", order.OrderNumber, humanStatus(order.Status))
	return s.send(ctx, to, name, subject, orderStatusHTML(name, order))
}

// SendSellerNewOrder notifies a seller that one of their items has been ordered.
func (s *EmailService) SendSellerNewOrder(ctx context.Context, to, name string, order *models.Order, items []models.OrderItem) error {
	if !s.enabled || to == "" || len(items) == 0 {
		return nil
	}
	subject := fmt.Sprintf("New order — %s", order.OrderNumber)
	return s.send(ctx, to, name, subject, sellerNewOrderHTML(name, order, items))
}

// SendPasswordReset emails a user a password-reset link containing a single-use token.
func (s *EmailService) SendPasswordReset(ctx context.Context, to, name, rawToken string) error {
	if !s.enabled || to == "" {
		return nil
	}
	link := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, rawToken)
	return s.send(ctx, to, name, "Reset your ArtShop password", passwordResetHTML(name, link))
}

// --- Brevo transport --------------------------------------------------------

type brevoContact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type brevoPayload struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
}

func (s *EmailService) send(ctx context.Context, to, toName, subject, html string) error {
	body, err := json.Marshal(brevoPayload{
		Sender:      brevoContact{Name: s.fromName, Email: s.fromEmail},
		To:          []brevoContact{{Name: toName, Email: to}},
		Subject:     subject,
		HTMLContent: html,
	})
	if err != nil {
		return fmt.Errorf("email_service: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brevoEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email_service: build request: %w", err)
	}
	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		slog.Error("email_service: HTTP error", "to", to, "subject", subject, "error", err)
		return fmt.Errorf("email_service: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		slog.Info("email_service: sent", "to", to, "subject", subject, "status", resp.StatusCode)
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	slog.Error("email_service: Brevo rejected request",
		"to", to, "subject", subject, "status", resp.StatusCode, "body", string(respBody))
	return fmt.Errorf("email_service: brevo returned %d: %s", resp.StatusCode, string(respBody))
}

func humanStatus(s string) string {
	switch s {
	case "pending":
		return "received"
	case "confirmed":
		return "confirmed"
	case "processing":
		return "being prepared"
	case "shipped":
		return "shipped"
	case "delivered":
		return "delivered"
	case "cancelled":
		return "cancelled"
	case "refunded":
		return "refunded"
	default:
		return s
	}
}

// --- HTML templates ---------------------------------------------------------

const baseStyles = `
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; color: #1f2937; background: #f9fafb; margin: 0; padding: 0; }
    .wrap { max-width: 560px; margin: 32px auto; padding: 0 16px; }
    .card { background: #ffffff; border: 1px solid #f1f5f9; border-radius: 16px; padding: 32px; }
    h1 { font-size: 22px; margin: 0 0 12px; color: #111827; }
    p { font-size: 15px; line-height: 1.6; color: #374151; margin: 0 0 14px; }
    .muted { color: #6b7280; font-size: 13px; }
    .badge { display: inline-block; background: #ecfdf5; color: #047857; padding: 4px 10px; border-radius: 999px; font-size: 12px; font-weight: 600; letter-spacing: 0.02em; }
    .total { font-size: 20px; font-weight: 700; color: #111827; }
    table { width: 100%; border-collapse: collapse; margin: 16px 0; }
    td { padding: 8px 0; font-size: 14px; vertical-align: top; }
    .item-name { color: #111827; }
    .right { text-align: right; }
    hr { border: none; border-top: 1px solid #f1f5f9; margin: 18px 0; }
    .button { display: inline-block; background: #111827; color: #ffffff !important; padding: 12px 28px; border-radius: 10px; font-weight: 600; font-size: 14px; text-decoration: none; margin: 8px 0; }
    .footer { text-align: center; color: #9ca3af; font-size: 12px; margin-top: 24px; }
  </style>
`

func welcomeHTML(name string) string {
	if name == "" {
		name = "there"
	}
	return fmt.Sprintf(`
<!doctype html><html><head><meta charset="utf-8">%s</head>
<body><div class="wrap"><div class="card">
  <span class="badge">WELCOME</span>
  <h1 style="margin-top:14px;">Welcome to ArtShop, %s!</h1>
  <p>Your account is ready. Browse one-of-a-kind handcrafted pieces from independent artists, save your favourites, and check out when you're ready.</p>
  <p class="muted">You're all set. Happy collecting.</p>
</div>
<p class="footer">ArtShop — handcrafted art, delivered.</p>
</div></body></html>`, baseStyles, name)
}

func orderConfirmationHTML(name string, o *models.Order) string {
	if name == "" {
		name = "there"
	}
	var rows strings.Builder
	for _, it := range o.Items {
		rows.WriteString(fmt.Sprintf(
			`<tr><td class="item-name">%s × %d</td><td class="right">$%.2f</td></tr>`,
			it.Title, it.Quantity, it.Price*float64(it.Quantity),
		))
	}

	addr := ""
	if o.ShippingAddressLine1 != nil {
		addr = *o.ShippingAddressLine1
		if o.ShippingCity != nil {
			addr += ", " + *o.ShippingCity
		}
		if o.ShippingCountry != nil {
			addr += ", " + *o.ShippingCountry
		}
	}

	return fmt.Sprintf(`
<!doctype html><html><head><meta charset="utf-8">%s</head>
<body><div class="wrap"><div class="card">
  <span class="badge">ORDER CONFIRMED</span>
  <h1 style="margin-top:14px;">Thanks, %s — your order is in.</h1>
  <p>We've received your order <strong>%s</strong>. Since this is a cash-on-delivery order, you'll pay when it arrives.</p>
  <hr/>
  <table>%s</table>
  <hr/>
  <table>
    <tr><td class="muted">Subtotal</td><td class="right">$%.2f</td></tr>
    <tr><td class="muted">Tax</td><td class="right">$%.2f</td></tr>
    <tr><td class="total">Total</td><td class="right total">$%.2f</td></tr>
  </table>
  <hr/>
  <p class="muted"><strong>Shipping to:</strong><br>%s</p>
  <p class="muted">Payment: <strong>Cash on Delivery</strong></p>
</div>
<p class="footer">ArtShop — handcrafted art, delivered.</p>
</div></body></html>`,
		baseStyles, name, o.OrderNumber, rows.String(),
		o.Subtotal, o.Tax, o.Total, addr,
	)
}

func orderStatusHTML(name string, o *models.Order) string {
	if name == "" {
		name = "there"
	}
	tracking := ""
	if o.TrackingNumber != nil && *o.TrackingNumber != "" {
		tracking = fmt.Sprintf(`<p class="muted">Tracking number: <strong>%s</strong></p>`, *o.TrackingNumber)
	}
	return fmt.Sprintf(`
<!doctype html><html><head><meta charset="utf-8">%s</head>
<body><div class="wrap"><div class="card">
  <span class="badge">ORDER UPDATE</span>
  <h1 style="margin-top:14px;">Hi %s, your order is %s.</h1>
  <p>Order <strong>%s</strong> is now <strong>%s</strong>.</p>
  %s
  <hr/>
  <p class="muted">Total: <strong>$%.2f</strong></p>
</div>
<p class="footer">ArtShop — handcrafted art, delivered.</p>
</div></body></html>`,
		baseStyles, name, humanStatus(o.Status), o.OrderNumber, humanStatus(o.Status),
		tracking, o.Total,
	)
}

func sellerNewOrderHTML(name string, o *models.Order, items []models.OrderItem) string {
	if name == "" {
		name = "Seller"
	}
	var rows strings.Builder
	var sellerTotal float64
	for _, it := range items {
		line := it.Price * float64(it.Quantity)
		sellerTotal += line
		rows.WriteString(fmt.Sprintf(
			`<tr><td class="item-name">%s × %d</td><td class="right">$%.2f</td></tr>`,
			it.Title, it.Quantity, line,
		))
	}
	return fmt.Sprintf(`
<!doctype html><html><head><meta charset="utf-8">%s</head>
<body><div class="wrap"><div class="card">
  <span class="badge">NEW ORDER</span>
  <h1 style="margin-top:14px;">Hi %s, you have a new order.</h1>
  <p>Order <strong>%s</strong> includes the following items from your shop:</p>
  <hr/>
  <table>%s</table>
  <hr/>
  <table>
    <tr><td class="total">Your subtotal</td><td class="right total">$%.2f</td></tr>
  </table>
  <p class="muted">Sign in to your seller dashboard to mark these as shipped.</p>
</div>
<p class="footer">ArtShop — handcrafted art, delivered.</p>
</div></body></html>`,
		baseStyles, name, o.OrderNumber, rows.String(), sellerTotal,
	)
}

func passwordResetHTML(name, link string) string {
	if name == "" {
		name = "there"
	}
	return fmt.Sprintf(`
<!doctype html><html><head><meta charset="utf-8">%s</head>
<body><div class="wrap"><div class="card">
  <span class="badge">PASSWORD RESET</span>
  <h1 style="margin-top:14px;">Reset your password, %s.</h1>
  <p>We received a request to reset the password on your ArtShop account. Click the button below to choose a new one. This link expires in 60 minutes and can only be used once.</p>
  <p style="text-align:center;"><a class="button" href="%s">Reset password</a></p>
  <p class="muted">If the button doesn't work, paste this URL into your browser:<br><span style="word-break:break-all;">%s</span></p>
  <hr/>
  <p class="muted">Didn't request this? You can safely ignore this email — your password won't be changed.</p>
</div>
<p class="footer">ArtShop — handcrafted art, delivered.</p>
</div></body></html>`, baseStyles, name, link, link)
}
