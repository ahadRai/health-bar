package middleware

import (
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters for different IPs
type IPRateLimiter struct {
    ips map[string]*rate.Limiter
    mu  *sync.RWMutex
    r   rate.Limit
    b   int
}

// NewIPRateLimiter creates a new IP rate limiter
// r = requests per second, b = burst size
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        ips: make(map[string]*rate.Limiter),
        mu:  &sync.RWMutex{},
        r:   r,
        b:   b,
    }
}

// GetLimiter returns the rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.ips[ip]
    if !exists {
        limiter = rate.NewLimiter(i.r, i.b)
        i.ips[ip] = limiter
    }

    return limiter
}

// CleanupOldEntries removes inactive limiters (run periodically)
func (i *IPRateLimiter) CleanupOldEntries() {
    i.mu.Lock()
    defer i.mu.Unlock()

    for ip, limiter := range i.ips {
        // Remove if no tokens have been used in the last hour
        if limiter.Tokens() == float64(i.b) {
            delete(i.ips, ip)
        }
    }
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *IPRateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Get client IP
            ip := getClientIP(r)
            
            // Get rate limiter for this IP
            rateLimiter := limiter.GetLimiter(ip)
            
            // Check if request is allowed
            if !rateLimiter.Allow() {
                http.Error(w, `{"success":false,"error":"Rate limit exceeded. Too many requests."}`, http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header (for proxies)
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return forwarded
    }

    // Check X-Real-IP header
    realIP := r.Header.Get("X-Real-IP")
    if realIP != "" {
        return realIP
    }

    // Fall back to RemoteAddr
    return r.RemoteAddr
}

// StartCleanup starts a goroutine to periodically clean up old entries
func (i *IPRateLimiter) StartCleanup(interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            i.CleanupOldEntries()
        }
    }()
}
