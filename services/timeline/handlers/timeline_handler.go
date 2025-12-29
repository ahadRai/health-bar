package handlers

import (
    "database/sql"
    "encoding/json"
    "health-bar/shared/models"
    "health-bar/shared/utils"
    "health-bar/services/timeline/repository"
    "net/http"
    "time"
)

type TimelineHandler struct {
    repo *repository.TimelineRepository
}

func NewTimelineHandler(repo *repository.TimelineRepository) *TimelineHandler {
    return &TimelineHandler{repo: repo}
}

type CreateVisitRequest struct {
    HospitalName string `json:"hospital_name"`
    VisitDate    string `json:"visit_date"` // Format: YYYY-MM-DD
    Reason       string `json:"reason"`
    Notes        string `json:"notes"`
}

type UpdateVisitRequest struct {
    HospitalName string `json:"hospital_name"`
    VisitDate    string `json:"visit_date"`
    Reason       string `json:"reason"`
    Notes        string `json:"notes"`
}

// CreateVisit creates a new hospital visit
func (h *TimelineHandler) CreateVisit(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can add hospital visits")
        return
    }

    // Get patient profile ID
    patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    var req CreateVisitRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.SendError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Validate
    if req.HospitalName == "" || req.VisitDate == "" || req.Reason == "" {
        utils.SendError(w, http.StatusBadRequest, "Hospital name, visit date, and reason are required")
        return
    }

    // Parse date
    visitDate, err := time.Parse("2006-01-02", req.VisitDate)
    if err != nil {
        utils.SendError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
        return
    }

    visit := &models.HospitalVisit{
        HospitalName: req.HospitalName,
        VisitDate:    visitDate,
        Reason:       req.Reason,
        Notes:        req.Notes,
    }

    if err := h.repo.CreateVisit(patientProfileID, visit); err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to create visit")
        return
    }

    utils.SendSuccess(w, http.StatusCreated, "Visit added successfully", visit)
}

// GetMyTimeline gets all hospital visits for the current patient
func (h *TimelineHandler) GetMyTimeline(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can view their timeline")
        return
    }

    // Get patient profile ID
    patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    visits, err := h.repo.GetVisitsByPatientID(patientProfileID)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve timeline")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Timeline retrieved", visits)
}

// GetPatientTimeline gets timeline for a specific patient (for doctors with access)
func (h *TimelineHandler) GetPatientTimeline(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // Get patient profile ID from query
    patientProfileID := r.URL.Query().Get("patient_id")
    if patientProfileID == "" {
        utils.SendError(w, http.StatusBadRequest, "Patient ID is required")
        return
    }

    // If patient is viewing their own timeline
    if userRole == "patient" {
        myProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
        if err != nil || myProfileID != patientProfileID {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else if userRole == "doctor" {
        // Check if doctor has access
        hasAccess, err := h.repo.CheckDoctorAccess(userID, patientProfileID)
        if err != nil || !hasAccess {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else {
        utils.SendError(w, http.StatusForbidden, "Access denied")
        return
    }

    visits, err := h.repo.GetVisitsByPatientID(patientProfileID)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve timeline")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Timeline retrieved", visits)
}

// GetVisit gets a specific hospital visit
func (h *TimelineHandler) GetVisit(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    visitID := r.URL.Query().Get("visit_id")
    if visitID == "" {
        utils.SendError(w, http.StatusBadRequest, "Visit ID is required")
        return
    }

    visit, err := h.repo.GetVisitByID(visitID)
    if err != nil {
        if err == sql.ErrNoRows {
            utils.SendError(w, http.StatusNotFound, "Visit not found")
            return
        }
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve visit")
        return
    }

    // Check if user has access
    if userRole == "patient" {
        patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
        if err != nil || visit.PatientID != patientProfileID {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else if userRole == "doctor" {
        hasAccess, err := h.repo.CheckDoctorAccess(userID, visit.PatientID)
        if err != nil || !hasAccess {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else {
        utils.SendError(w, http.StatusForbidden, "Access denied")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Visit retrieved", visit)
}

// UpdateVisit updates a hospital visit
func (h *TimelineHandler) UpdateVisit(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can update visits")
        return
    }

    visitID := r.URL.Query().Get("visit_id")
    if visitID == "" {
        utils.SendError(w, http.StatusBadRequest, "Visit ID is required")
        return
    }

    // Check if visit belongs to this patient
    patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    visitPatientID, err := h.repo.GetPatientIDByVisitID(visitID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Visit not found")
        return
    }

    if visitPatientID != patientProfileID {
        utils.SendError(w, http.StatusForbidden, "Access denied")
        return
    }

    var req UpdateVisitRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.SendError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Parse date
    visitDate, err := time.Parse("2006-01-02", req.VisitDate)
    if err != nil {
        utils.SendError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
        return
    }

    visit := &models.HospitalVisit{
        HospitalName: req.HospitalName,
        VisitDate:    visitDate,
        Reason:       req.Reason,
        Notes:        req.Notes,
    }

    if err := h.repo.UpdateVisit(visitID, visit); err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to update visit")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Visit updated successfully", visit)
}

// DeleteVisit deletes a hospital visit
func (h *TimelineHandler) DeleteVisit(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can delete visits")
        return
    }

    visitID := r.URL.Query().Get("visit_id")
    if visitID == "" {
        utils.SendError(w, http.StatusBadRequest, "Visit ID is required")
        return
    }

    // Check if visit belongs to this patient
    patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    visitPatientID, err := h.repo.GetPatientIDByVisitID(visitID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Visit not found")
        return
    }

    if visitPatientID != patientProfileID {
        utils.SendError(w, http.StatusForbidden, "Access denied")
        return
    }

    if err := h.repo.DeleteVisit(visitID); err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to delete visit")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Visit deleted successfully", nil)
}
