package handlers

import (
	"database/sql"
	"encoding/json"
	"health-bar/services/patient/repository"
	"health-bar/shared/models"
	"health-bar/shared/utils"
	"net/http"
	"strings"
	"time"
)

type PatientHandler struct {
	repo *repository.PatientRepository
}

func NewPatientHandler(repo *repository.PatientRepository) *PatientHandler {
	return &PatientHandler{repo: repo}
}

type CreateProfileRequest struct {
	FullName    string `json:"full_name"`
	DateOfBirth string `json:"date_of_birth"` // Format: YYYY-MM-DD
	Gender      string `json:"gender"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
}

type UpdateProfileRequest struct {
	FullName    string `json:"full_name"`
	DateOfBirth string `json:"date_of_birth"`
	Gender      string `json:"gender"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
}

// CreateProfile creates a patient profile
func (h *PatientHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	if req.FullName == "" || req.DateOfBirth == "" {
		utils.SendError(w, http.StatusBadRequest, "Full name and date of birth are required")
		return
	}

	// Parse date
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	profile := &models.PatientProfile{
		FullName:    req.FullName,
		DateOfBirth: dob,
		Gender:      req.Gender,
		Phone:       req.Phone,
		Address:     req.Address,
	}

	if err := h.repo.CreateProfile(userID, profile); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			utils.SendError(w, http.StatusConflict, "Profile already exists")
			return
		}
		utils.SendError(w, http.StatusInternalServerError, "Failed to create profile")
		return
	}

	utils.SendSuccess(w, http.StatusCreated, "Profile created successfully", profile)
}

// GetMyProfile gets the current user's patient profile
func (h *PatientHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := h.repo.GetProfileByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendError(w, http.StatusNotFound, "Profile not found")
			return
		}
		utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "Profile retrieved", profile)
}

// UpdateProfile updates patient profile
func (h *PatientHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse date
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	profile := &models.PatientProfile{
		FullName:    req.FullName,
		DateOfBirth: dob,
		Gender:      req.Gender,
		Phone:       req.Phone,
		Address:     req.Address,
	}

	if err := h.repo.UpdateProfile(userID, profile); err != nil {
		if err == sql.ErrNoRows {
			utils.SendError(w, http.StatusNotFound, "Profile not found")
			return
		}
		utils.SendError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "Profile updated successfully", profile)
}

// GenerateShareLink generates a token for sharing the profile
func (h *PatientHandler) GenerateShareLink(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get patient profile ID
	profile, err := h.repo.GetProfileByUserID(userID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Patient profile not found")
		return
	}

	token, err := h.repo.GenerateShareToken(profile.ID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to generate share token")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "Share token generated", map[string]string{
		"share_token": token,
		"share_url":   "/api/patients/share/" + token, // Frontend can construct full URL
	})
}

// GetSharedProfile retrieves a profile using the share token (public access)
func (h *PatientHandler) GetSharedProfile(w http.ResponseWriter, r *http.Request) {
	// Get token from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) == 0 {
		utils.SendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	token := pathParts[len(pathParts)-1]

	if token == "" {
		utils.SendError(w, http.StatusBadRequest, "Share token is required")
		return
	}

	profile, err := h.repo.GetProfileByShareToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendError(w, http.StatusNotFound, "Profile not found or invalid token")
			return
		}
		utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}

	// Return limited info if needed, but for now returning full profile
	utils.SendSuccess(w, http.StatusOK, "Profile retrieved", profile)
}
