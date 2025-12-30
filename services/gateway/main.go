package main

import (
    "health-bar/services/gateway/handlers"
    "health-bar/services/gateway/middleware"
    "log"
    "net/http"
    "os"
    "time"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
    "golang.org/x/time/rate"
)

func main() {
    godotenv.Load()

    // Service configuration
    config := handlers.ServiceConfig{
        AuthServiceURL:        getEnv("AUTH_SERVICE_URL", "http://healthbar-auth-service:8001"),
        PatientServiceURL:     getEnv("PATIENT_SERVICE_URL", "http://healthbar-patient-service:8002"),
        DoctorServiceURL:      getEnv("DOCTOR_SERVICE_URL", "http://healthbar-doctor-service:8003"),
        TimelineServiceURL:    getEnv("TIMELINE_SERVICE_URL", "http://healthbar-timeline-service:8004"),
        PrescriptionServiceURL: getEnv("PRESCRIPTION_SERVICE_URL", "http://healthbar-prescription-service:8005"),
    }

    proxyHandler := handlers.NewProxyHandler(config)

    // Create rate limiter (10 requests per second, burst of 20)
    rateLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20)
    
    // Start cleanup goroutine (clean every 5 minutes)
    rateLimiter.StartCleanup(5 * time.Minute)

    // Create router
    router := mux.NewRouter()

    // Health check endpoint (no rate limit)
    router.HandleFunc("/health", proxyHandler.HealthCheck).Methods("GET")

    // API Gateway info
    router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{
            "service": "Health Bar API Gateway",
            "version": "1.0.0",
            "endpoints": {
                "auth": "/api/auth/*",
                "patients": "/api/patients/*",
                "doctors": "/api/doctors/*",
                "timeline": "/api/timeline/*",
                "prescriptions": "/api/prescriptions/*"
            },
            "rate_limit": "10 requests per second, burst 20"
        }`))
    }).Methods("GET")

    // All API routes go through proxy with rate limiting
    apiRouter := router.PathPrefix("/api").Subrouter()
    apiRouter.PathPrefix("/").HandlerFunc(proxyHandler.ProxyRequest)

    // Apply middlewares
    handler := middleware.LoggingMiddleware(router)
    handler = middleware.RateLimitMiddleware(rateLimiter)(handler)

    // CORS configuration
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    port := getEnv("PORT", "8000")
    log.Printf("API Gateway starting on port %s", port)
    log.Printf("Rate limit: 10 requests/second per IP, burst: 20")
    log.Printf("Routing:")
    log.Printf("  /api/auth/*         -> %s", config.AuthServiceURL)
    log.Printf("  /api/patients/*     -> %s", config.PatientServiceURL)
    log.Printf("  /api/doctors/*      -> %s", config.DoctorServiceURL)
    log.Printf("  /api/timeline/*     -> %s", config.TimelineServiceURL)
    log.Printf("  /api/prescriptions/* -> %s", config.PrescriptionServiceURL)
    
    log.Fatal(http.ListenAndServe(":"+port, c.Handler(handler)))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
