package main

import (
    "health-bar/shared/database"
    "health-bar/shared/middleware"
    "health-bar/services/timeline/handlers"
    "health-bar/services/timeline/repository"
    "log"
    "net/http"
    "os"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
)

func main() {
    godotenv.Load()

    db, err := database.Connect(database.Config{
        Host:     getEnv("DB_HOST", "localhost"),
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

    repo := repository.NewTimelineRepository(db)
    handler := handlers.NewTimelineHandler(repo)

    router := mux.NewRouter()

    // Timeline routes (protected)
    router.HandleFunc("/api/timeline/visits", middleware.AuthMiddleware(handler.CreateVisit)).Methods("POST")
    router.HandleFunc("/api/timeline/my", middleware.AuthMiddleware(handler.GetMyTimeline)).Methods("GET")
    router.HandleFunc("/api/timeline/patient", middleware.AuthMiddleware(handler.GetPatientTimeline)).Methods("GET")
    router.HandleFunc("/api/timeline/visit", middleware.AuthMiddleware(handler.GetVisit)).Methods("GET")
    router.HandleFunc("/api/timeline/visit", middleware.AuthMiddleware(handler.UpdateVisit)).Methods("PUT")
    router.HandleFunc("/api/timeline/visit", middleware.AuthMiddleware(handler.DeleteVisit)).Methods("DELETE")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    port := getEnv("PORT", "8004")
    log.Printf("Timeline service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, c.Handler(router)))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
