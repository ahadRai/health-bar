package main

import (
    "health-bar/shared/database"
    "health-bar/shared/middleware"
    "health-bar/services/prescription/handlers"
    "health-bar/services/prescription/repository"
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

    // Set upload path
    uploadPath := getEnv("UPLOAD_PATH", "/app/uploads")

    repo := repository.NewPrescriptionRepository(db)
    handler := handlers.NewPrescriptionHandler(repo, uploadPath)

    router := mux.NewRouter()

    // Prescription routes (protected)
    router.HandleFunc("/api/prescriptions/upload", middleware.AuthMiddleware(handler.UploadPrescription)).Methods("POST")
    router.HandleFunc("/api/prescriptions/my", middleware.AuthMiddleware(handler.GetMyPrescriptions)).Methods("GET")
    router.HandleFunc("/api/prescriptions/patient", middleware.AuthMiddleware(handler.GetPatientPrescriptions)).Methods("GET")
    router.HandleFunc("/api/prescriptions/download", middleware.AuthMiddleware(handler.DownloadPrescription)).Methods("GET")
    router.HandleFunc("/api/prescriptions", middleware.AuthMiddleware(handler.DeletePrescription)).Methods("DELETE")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    port := getEnv("PORT", "8005")
    log.Printf("Prescription service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, c.Handler(router)))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
