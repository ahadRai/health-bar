package handlers

import (
    "database/sql"
    "encoding/json"
    "health-bar/shared/models"
    "health-bar/shared/utils"
    "health-bar/services/doctor/repository"
    "net/http"
    "strings"
)

type DoctorHandler struct {
    repo *repository.DoctorRepository
}

func NewDoctorHandler(repo *repository.DoctorRepository) *DoctorHandler {
    return &DoctorHandler{repo: repo}
}

type CreateProfileRequest struct {
    FullName       string `json:"full_name"`
    Specialization string `json:"specialization"`
    LicenseNumber  string `json:"license_number"`
    Phone          string `json:"phone"`
}

type UpdateProfileRequest struct {
    FullName       string `json:"full_name"`
    Specialization string `json:"specialization"`
    LicenseNumber  string `json:"license_number"`
    Phone          string `json:"phone"`
}

// CreateProfile creates a doctor profile
func (h *DoctorHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "doctor" {
        utils.SendError(w, http.StatusForbidden, "Only doctors can create doctor profiles")
        return
    }

    var req CreateProfileRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.SendError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Validate
    if req.FullName == "" {
        utils.SendError(w, http.StatusBadRequest, "Full name is required")
        return
    }

    profile := &models.DoctorProfile{
        FullName:       req.FullName,
        Specialization: req.Specialization,
        LicenseNumber:  req.LicenseNumber,
        Phone:          req.Phone,
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

// GetMyProfile gets the current doctor's profile
func (h *DoctorHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "doctor" {
        utils.SendError(w, http.StatusForbidden, "Only doctors can view their profile")
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

// UpdateProfile updates doctor profile
func (h *DoctorHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "doctor" {
        utils.SendError(w, http.StatusForbidden, "Only doctors can update their profile")
        return
    }

    var req UpdateProfileRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.SendError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    profile := &models.DoctorProfile{
        FullName:       req.FullName,
        Specialization: req.Specialization,
        LicenseNumber:  req.LicenseNumber,
        Phone:          req.Phone,
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

// GetPatientProfile gets a patient's profile (if doctor has access)
func (h *DoctorHandler) GetPatientProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "doctor" {
        utils.SendError(w, http.StatusForbidden, "Only doctors can view patient profiles")
        return
    }

    // Get doctor profile
    doctorProfile, err := h.repo.GetProfileByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Doctor profile not found")
        return
    }

    // Get patient ID from URL query
    patientID := r.URL.Query().Get("patient_id")
    if patientID == "" {
        utils.SendError(w, http.StatusBadRequest, "Patient ID is required")
        return
    }

    // Get patient profile (with permission check)
    patientProfile, err := h.repo.GetPatientProfile(doctorProfile.ID, patientID)
    if err != nil {
        if err == sql.ErrNoRows {
            utils.SendError(w, http.StatusForbidden, "Access denied or patient not found")
            return
        }
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve patient profile")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Patient profile retrieved", patientProfile)
}

// ListAccessiblePatients lists all patients the doctor can access
func (h *DoctorHandler) ListAccessiblePatients(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "doctor" {
        utils.SendError(w, http.StatusForbidden, "Only doctors can view patient list")
        return
    }

    // Get doctor profile
    doctorProfile, err := h.repo.GetProfileByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Doctor profile not found")
        return
    }

    // Get accessible patients
    patients, err := h.repo.ListAccessiblePatients(doctorProfile.ID)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve patients")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Patients retrieved", patients)
}
