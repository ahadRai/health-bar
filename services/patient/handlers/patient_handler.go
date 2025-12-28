package handlers

import (
    "database/sql"
    "encoding/json"
    "health-bar/shared/models"
    "health-bar/shared/utils"
    "health-bar/services/patient/repository"
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
    FullName    string    `json:"full_name"`
    DateOfBirth string    `json:"date_of_birth"` // Format: YYYY-MM-DD
    Gender      string    `json:"gender"`
    Phone       string    `json:"phone"`
    Address     string    `json:"address"`
}

type UpdateProfileRequest struct {
    FullName    string `json:"full_name"`
    DateOfBirth string `json:"date_of_birth"`
    Gender      string `json:"gender"`
    Phone       string `json:"phone"`
    Address     string `json:"address"`
}

type GrantAccessRequest struct {
    DoctorID string `json:"doctor_id"`
}

// CreateProfile creates a patient profile
func (h *PatientHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can create patient profiles")
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
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can view their profile")
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
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can update their profile")
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

// GrantAccess grants a doctor access to patient's records
func (h *PatientHandler) GrantAccess(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can grant access")
        return
    }

    // Get patient profile ID
    profile, err := h.repo.GetProfileByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    var req GrantAccessRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.SendError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    if req.DoctorID == "" {
        utils.SendError(w, http.StatusBadRequest, "Doctor ID is required")
        return
    }

    if err := h.repo.GrantAccess(profile.ID, req.DoctorID); err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to grant access")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Access granted successfully", nil)
}

// RevokeAccess revokes a doctor's access
func (h *PatientHandler) RevokeAccess(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can revoke access")
        return
    }

    // Get patient profile ID
    profile, err := h.repo.GetProfileByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    // Get doctor ID from URL
    doctorID := r.URL.Query().Get("doctor_id")
    if doctorID == "" {
        utils.SendError(w, http.StatusBadRequest, "Doctor ID is required")
        return
    }

    if err := h.repo.RevokeAccess(profile.ID, doctorID); err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to revoke access")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Access revoked successfully", nil)
}

// ListPermissions lists all access permissions
func (h *PatientHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can view permissions")
        return
    }

    // Get patient profile ID
    profile, err := h.repo.GetProfileByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    permissions, err := h.repo.ListPermissions(profile.ID)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve permissions")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Permissions retrieved", permissions)
}
