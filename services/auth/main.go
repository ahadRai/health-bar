package main

import (
    "health-bar/database"
    "health-bar/services/auth/handlers"
    "health-bar/services/auth/repository"
    "log"
    "net/http"
    "os"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
)

func main() {
    // Load environment variables
    godotenv.Load()

    // Database connection
    db, err := database.Connect(database.Config{
        Host:     getEnv("DB_HOST", "172.17.0.2"),
        Port:     getEnv("DB_PORT", "5432"),
        User:     getEnv("DB_USER", "postgres"),
        Password: getEnv("DB_PASSWORD", "postgres"),
        DBName:   getEnv("DB_NAME", "healthbar"),
        SSLMode:  getEnv("DB_SSLMODE", "disable"),
    })
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Initialize repository and handlers
    repo := repository.NewAuthRepository(db)
    handler := handlers.NewAuthHandler(repo)

    // Setup router
    router := mux.NewRouter()
    
    // Auth routes
    router.HandleFunc("/api/auth/register", handler.Register).Methods("POST")
    router.HandleFunc("/api/auth/login", handler.Login).Methods("POST")
    router.HandleFunc("/api/auth/me", handler.GetCurrentUser).Methods("GET")

    // CORS
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    // Start server
    port := getEnv("PORT", "8001")
    log.Printf("Auth service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, c.Handler(router)))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}