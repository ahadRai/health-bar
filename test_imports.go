package main

import (
    "fmt"
    "health-bar/shared/utils"
    "health-bar/shared/models"
)

func main() {
    // Test JWT
    token, err := utils.GenerateToken("123", "test@example.com", "patient")
    if err != nil {
        fmt.Println("JWT Error:", err)
        return
    }
    fmt.Println("JWT Token generated:", token[:20]+"...")
    
    // Test password hashing
    hash, _ := utils.HashPassword("password123")
    fmt.Println("Password hashed successfully:", hash[:20]+"...")
    
    // Test models
    user := models.User{Email: "test@example.com"}
    fmt.Println("User model created:", user.Email)
    
    fmt.Println("\n✅ All imports working!")
}
