package middleware

import (
    "log"
    "net/http"
    "time"
)

// LoggingMiddleware logs all requests
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Create a response writer wrapper to capture status code
        wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

        // Process request
        next.ServeHTTP(wrapped, r)

        // Log request details
        duration := time.Since(start)
        log.Printf(
            "%s %s %s %d %v",
            r.Method,
            r.RequestURI,
            r.RemoteAddr,
            wrapped.statusCode,
            duration,
        )
    })
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}
