package middleware

import (
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/artshop/backend/pkg/response"
)

// tokenBucket implements a simple token-bucket rate limiter for a single client.
type tokenBucket struct {
	tokens   float64
	maxBurst float64
	rps      float64 // tokens added per second
	lastSeen time.Time
}

// allow checks whether the bucket has at least one token available. If so it
// consumes a token and returns true; otherwise it returns false.
func (b *tokenBucket) allow(now time.Time) bool {
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens = math.Min(b.maxBurst, b.tokens+elapsed*b.rps)
	b.lastSeen = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

// RateLimiter is a per-IP token-bucket rate limiter. It stores buckets in a
// sync.Map and periodically evicts stale entries to prevent unbounded growth.
type RateLimiter struct {
	rps   float64
	burst int

	mu      sync.Mutex
	buckets sync.Map // map[string]*tokenBucket

	// cleanupInterval controls how often stale entries are removed.
	cleanupInterval time.Duration
	// staleAfter is the duration after which an unseen IP is evicted.
	staleAfter time.Duration
	stopCh     chan struct{}
}

// NewRateLimiter creates a RateLimiter with the given requests-per-second and
// burst size, and starts a background goroutine for periodic cleanup.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		rps:             rps,
		burst:           burst,
		cleanupInterval: 1 * time.Minute,
		staleAfter:      3 * time.Minute,
		stopCh:          make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

// Stop terminates the background cleanup goroutine. Call this on shutdown.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// Middleware returns a chi-compatible HTTP middleware that enforces the rate
// limit per client IP. When a client exceeds the limit it receives a 429
// status with a JSON error body.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)

		now := time.Now()

		// Load or create a bucket for this IP.
		val, _ := rl.buckets.LoadOrStore(ip, &tokenBucket{
			tokens:   float64(rl.burst),
			maxBurst: float64(rl.burst),
			rps:      rl.rps,
			lastSeen: now,
		})
		bucket := val.(*tokenBucket)

		rl.mu.Lock()
		allowed := bucket.allow(now)
		rl.mu.Unlock()

		if !allowed {
			w.Header().Set("Retry-After", "1")
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Please try again later.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// cleanupLoop periodically removes stale IP entries from the bucket map.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.evictStale()
		}
	}
}

// evictStale removes entries that have not been seen within the stale window.
func (rl *RateLimiter) evictStale() {
	cutoff := time.Now().Add(-rl.staleAfter)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.buckets.Range(func(key, value any) bool {
		bucket := value.(*tokenBucket)
		if bucket.lastSeen.Before(cutoff) {
			rl.buckets.Delete(key)
		}
		return true
	})
}

// realIP attempts to determine the real client IP, respecting common proxy
// headers. Falls back to the connection's remote address.
func realIP(r *http.Request) string {
	// X-Real-IP is commonly set by reverse proxies like nginx.
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// X-Forwarded-For may contain a comma-separated list; the first entry is
	// the original client.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := len(xff); i > 0 {
			for idx := 0; idx < len(xff); idx++ {
				if xff[idx] == ',' {
					return xff[:idx]
				}
			}
			return xff
		}
	}

	// Strip the port from RemoteAddr.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
