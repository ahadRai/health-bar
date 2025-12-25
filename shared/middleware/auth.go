package middleware

import (
    "health-bar/shared/utils"
    "net/http"
    "strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }

        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
        claims, err := utils.ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add claims to request context
        r.Header.Set("X-User-ID", claims.UserID)
        r.Header.Set("X-User-Email", claims.Email)
        r.Header.Set("X-User-Role", claims.Role)

        next(w, r)
    }
}