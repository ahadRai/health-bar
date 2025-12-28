package main

import (
    "health-bar/shared/database"
    "health-bar/shared/middleware"
    "health-bar/services/patient/handlers"
    "health-bar/services/patient/repository"
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

    repo := repository.NewPatientRepository(db)
    handler := handlers.NewPatientHandler(repo)

    router := mux.NewRouter()

    // Patient profile routes (protected)
    router.HandleFunc("/api/patients/profile", middleware.AuthMiddleware(handler.CreateProfile)).Methods("POST")
    router.HandleFunc("/api/patients/profile", middleware.AuthMiddleware(handler.GetMyProfile)).Methods("GET")
    router.HandleFunc("/api/patients/profile", middleware.AuthMiddleware(handler.UpdateProfile)).Methods("PUT")

    // Access permission routes (protected)
    router.HandleFunc("/api/patients/permissions/grant", middleware.AuthMiddleware(handler.GrantAccess)).Methods("POST")
    router.HandleFunc("/api/patients/permissions/revoke", middleware.AuthMiddleware(handler.RevokeAccess)).Methods("DELETE")
    router.HandleFunc("/api/patients/permissions", middleware.AuthMiddleware(handler.ListPermissions)).Methods("GET")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    port := getEnv("PORT", "8002")
    log.Printf("Patient service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, c.Handler(router)))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
