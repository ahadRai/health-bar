package main

import (
    "health-bar/shared/database"
    "health-bar/shared/middleware"
    "health-bar/services/doctor/handlers"
    "health-bar/services/doctor/repository"
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

    repo := repository.NewDoctorRepository(db)
    handler := handlers.NewDoctorHandler(repo)

    router := mux.NewRouter()

    // Doctor profile routes (protected)
    router.HandleFunc("/api/doctors/profile", middleware.AuthMiddleware(handler.CreateProfile)).Methods("POST")
    router.HandleFunc("/api/doctors/profile", middleware.AuthMiddleware(handler.GetMyProfile)).Methods("GET")
    router.HandleFunc("/api/doctors/profile", middleware.AuthMiddleware(handler.UpdateProfile)).Methods("PUT")

    // Patient viewing routes (protected)
    router.HandleFunc("/api/doctors/patients", middleware.AuthMiddleware(handler.ListAccessiblePatients)).Methods("GET")
    router.HandleFunc("/api/doctors/patients/view", middleware.AuthMiddleware(handler.GetPatientProfile)).Methods("GET")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    port := getEnv("PORT", "8003")
    log.Printf("Doctor service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, c.Handler(router)))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
