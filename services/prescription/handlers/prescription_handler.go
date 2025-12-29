package handlers

import (
    "database/sql"
    "fmt"
    "health-bar/shared/models"
    "health-bar/shared/utils"
    "health-bar/services/prescription/repository"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type PrescriptionHandler struct {
    repo       *repository.PrescriptionRepository
    uploadPath string
}

func NewPrescriptionHandler(repo *repository.PrescriptionRepository, uploadPath string) *PrescriptionHandler {
    // Create upload directory if it doesn't exist
    os.MkdirAll(uploadPath, 0755)
    return &PrescriptionHandler{
        repo:       repo,
        uploadPath: uploadPath,
    }
}

// UploadPrescription handles file upload
func (h *PrescriptionHandler) UploadPrescription(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can upload prescriptions")
        return
    }

    // Get patient profile ID
    patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    // Parse multipart form (max 10MB)
    err = r.ParseMultipartForm(10 << 20) // 10 MB
    if err != nil {
        utils.SendError(w, http.StatusBadRequest, "File too large. Max size is 10MB")
        return
    }

    // Get file from form
    file, header, err := r.FormFile("file")
    if err != nil {
        utils.SendError(w, http.StatusBadRequest, "No file uploaded")
        return
    }
    defer file.Close()

    // Validate file type (PDF or images)
    fileExt := strings.ToLower(filepath.Ext(header.Filename))
    allowedTypes := map[string]bool{
        ".pdf":  true,
        ".jpg":  true,
        ".jpeg": true,
        ".png":  true,
    }

    if !allowedTypes[fileExt] {
        utils.SendError(w, http.StatusBadRequest, "Invalid file type. Only PDF, JPG, JPEG, and PNG are allowed")
        return
    }

    // Generate unique filename
    uniqueFilename := fmt.Sprintf("%s_%s%s", patientProfileID, utils.GenerateUUID(), fileExt)
    filePath := filepath.Join(h.uploadPath, uniqueFilename)

    // Create file on disk
    dst, err := os.Create(filePath)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to save file")
        return
    }
    defer dst.Close()

    // Copy uploaded file to destination
    fileSize, err := io.Copy(dst, file)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to save file")
        return
    }

    // Save to database
    prescription := &models.Prescription{
        FileName: header.Filename,
        FileType: fileExt,
        FileSize: fileSize,
        FilePath: uniqueFilename, // Store only filename, not full path
    }

    if err := h.repo.CreatePrescription(patientProfileID, prescription); err != nil {
        // Delete file if database insert fails
        os.Remove(filePath)
        utils.SendError(w, http.StatusInternalServerError, "Failed to save prescription record")
        return
    }

    utils.SendSuccess(w, http.StatusCreated, "Prescription uploaded successfully", prescription)
}

// GetMyPrescriptions gets all prescriptions for the current patient
func (h *PrescriptionHandler) GetMyPrescriptions(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can view their prescriptions")
        return
    }

    // Get patient profile ID
    patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Patient profile not found")
        return
    }

    prescriptions, err := h.repo.GetPrescriptionsByPatientID(patientProfileID)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve prescriptions")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Prescriptions retrieved", prescriptions)
}

// GetPatientPrescriptions gets prescriptions for a specific patient (for doctors with access)
func (h *PrescriptionHandler) GetPatientPrescriptions(w http.ResponseWriter, r *http.Request) {
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

    // Check access
    if userRole == "patient" {
        myProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
        if err != nil || myProfileID != patientProfileID {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else if userRole == "doctor" {
        hasAccess, err := h.repo.CheckDoctorAccess(userID, patientProfileID)
        if err != nil || !hasAccess {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else {
        utils.SendError(w, http.StatusForbidden, "Access denied")
        return
    }

    prescriptions, err := h.repo.GetPrescriptionsByPatientID(patientProfileID)
    if err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve prescriptions")
        return
    }

    utils.SendSuccess(w, http.StatusOK, "Prescriptions retrieved", prescriptions)
}

// DownloadPrescription downloads a prescription file
func (h *PrescriptionHandler) DownloadPrescription(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    prescriptionID := r.URL.Query().Get("id")
    if prescriptionID == "" {
        utils.SendError(w, http.StatusBadRequest, "Prescription ID is required")
        return
    }

    // Get prescription
    prescription, err := h.repo.GetPrescriptionByID(prescriptionID)
    if err != nil {
        if err == sql.ErrNoRows {
            utils.SendError(w, http.StatusNotFound, "Prescription not found")
            return
        }
        utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve prescription")
        return
    }

    // Check access
    if userRole == "patient" {
        patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
        if err != nil || prescription.PatientID != patientProfileID {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else if userRole == "doctor" {
        hasAccess, err := h.repo.CheckDoctorAccess(userID, prescription.PatientID)
        if err != nil || !hasAccess {
            utils.SendError(w, http.StatusForbidden, "Access denied")
            return
        }
    } else {
        utils.SendError(w, http.StatusForbidden, "Access denied")
        return
    }

    // Open file
    filePath := filepath.Join(h.uploadPath, prescription.FilePath)
    file, err := os.Open(filePath)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "File not found")
        return
    }
    defer file.Close()

    // Set headers for file download
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", prescription.FileName))
    w.Header().Set("Content-Type", getContentType(prescription.FileType))

    // Stream file to response
    io.Copy(w, file)
}

// DeletePrescription deletes a prescription
func (h *PrescriptionHandler) DeletePrescription(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    userRole := r.Header.Get("X-User-Role")

    if userID == "" {
        utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    if userRole != "patient" {
        utils.SendError(w, http.StatusForbidden, "Only patients can delete prescriptions")
        return
    }

    prescriptionID := r.URL.Query().Get("id")
    if prescriptionID == "" {
        utils.SendError(w, http.StatusBadRequest, "Prescription ID is required")
        return
    }

    // Get prescription
    prescription, err := h.repo.GetPrescriptionByID(prescriptionID)
    if err != nil {
        utils.SendError(w, http.StatusNotFound, "Prescription not found")
        return
    }

    // Check if belongs to patient
    patientProfileID, err := h.repo.GetPatientProfileIDByUserID(userID)
    if err != nil || prescription.PatientID != patientProfileID {
        utils.SendError(w, http.StatusForbidden, "Access denied")
        return
    }

    // Delete from database
    if err := h.repo.DeletePrescription(prescriptionID); err != nil {
        utils.SendError(w, http.StatusInternalServerError, "Failed to delete prescription")
        return
    }

    // Delete file from disk
    filePath := filepath.Join(h.uploadPath, prescription.FilePath)
    os.Remove(filePath) // Ignore error if file doesn't exist

    utils.SendSuccess(w, http.StatusOK, "Prescription deleted successfully", nil)
}

// Helper function to get content type
func getContentType(fileType string) string {
    switch fileType {
    case ".pdf":
        return "application/pdf"
    case ".jpg", ".jpeg":
        return "image/jpeg"
    case ".png":
        return "image/png"
    default:
        return "application/octet-stream"
    }
}
