package handlers

import (
	"encoding/json"
	"health-bar/services/auth/repository"
	"health-bar/shared/models"
	"health-bar/shared/utils"
	"net/http"
	"strings"
)

type AuthHandler struct {
	repo *repository.AuthRepository
}

func NewAuthHandler(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type RegisterRequest struct {
	Email    string          `json:"email"`
	Password string          `json:"password"`
	Role     models.UserRole `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	if req.Email == "" || req.Password == "" {
		utils.SendError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if req.Role != models.RolePatient && req.Role != models.RoleDoctor {
		utils.SendError(w, http.StatusBadRequest, "Role must be 'patient' or 'doctor'")
		return
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	user, err := h.repo.CreateUser(req.Email, passwordHash, req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			utils.SendError(w, http.StatusConflict, "Email already exists")
			return
		}
		utils.SendError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SendSuccess(w, http.StatusCreated, "User registered successfully", AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user
	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		utils.SendError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "Login successful", AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.SendError(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Get user
	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User retrieved", user)
}
