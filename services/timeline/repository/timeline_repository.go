package repository

import (
    "health-bar/shared/models"
    "github.com/jmoiron/sqlx"
    "github.com/google/uuid"
)

type TimelineRepository struct {
    db *sqlx.DB
}

func NewTimelineRepository(db *sqlx.DB) *TimelineRepository {
    return &TimelineRepository{db: db}
}

// CreateVisit creates a new hospital visit
func (r *TimelineRepository) CreateVisit(patientID string, visit *models.HospitalVisit) error {
    visit.ID = uuid.New().String()
    visit.PatientID = patientID

    query := `
        INSERT INTO hospital_visits (id, patient_id, hospital_name, visit_date, reason, notes)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, patient_id, hospital_name, visit_date, reason, notes, created_at, updated_at
    `

    return r.db.QueryRowx(query,
        visit.ID, visit.PatientID, visit.HospitalName, visit.VisitDate,
        visit.Reason, visit.Notes,
    ).StructScan(visit)
}

// GetVisitByID gets a hospital visit by ID
func (r *TimelineRepository) GetVisitByID(visitID string) (*models.HospitalVisit, error) {
    visit := &models.HospitalVisit{}
    query := `
        SELECT id, patient_id, hospital_name, visit_date, reason, notes, created_at, updated_at
        FROM hospital_visits
        WHERE id = $1
    `
    err := r.db.Get(visit, query, visitID)
    return visit, err
}

// GetVisitsByPatientID gets all visits for a patient
func (r *TimelineRepository) GetVisitsByPatientID(patientID string) ([]models.HospitalVisit, error) {
    var visits []models.HospitalVisit
    query := `
        SELECT id, patient_id, hospital_name, visit_date, reason, notes, created_at, updated_at
        FROM hospital_visits
        WHERE patient_id = $1
        ORDER BY visit_date DESC, created_at DESC
    `
    err := r.db.Select(&visits, query, patientID)
    return visits, err
}

// UpdateVisit updates a hospital visit
func (r *TimelineRepository) UpdateVisit(visitID string, visit *models.HospitalVisit) error {
    query := `
        UPDATE hospital_visits
        SET hospital_name = $1, visit_date = $2, reason = $3, notes = $4, updated_at = NOW()
        WHERE id = $5
        RETURNING id, patient_id, hospital_name, visit_date, reason, notes, created_at, updated_at
    `

    return r.db.QueryRowx(query,
        visit.HospitalName, visit.VisitDate, visit.Reason, visit.Notes, visitID,
    ).StructScan(visit)
}

// DeleteVisit deletes a hospital visit
func (r *TimelineRepository) DeleteVisit(visitID string) error {
    query := `DELETE FROM hospital_visits WHERE id = $1`
    _, err := r.db.Exec(query, visitID)
    return err
}

// GetPatientIDByVisitID gets the patient ID associated with a visit
func (r *TimelineRepository) GetPatientIDByVisitID(visitID string) (string, error) {
    var patientID string
    query := `SELECT patient_id FROM hospital_visits WHERE id = $1`
    err := r.db.Get(&patientID, query, visitID)
    return patientID, err
}

// GetPatientProfileIDByUserID gets patient profile ID from user ID
func (r *TimelineRepository) GetPatientProfileIDByUserID(userID string) (string, error) {
    var profileID string
    query := `SELECT id FROM patient_profiles WHERE user_id = $1`
    err := r.db.Get(&profileID, query, userID)
    return profileID, err
}

// CheckDoctorAccess checks if a doctor has access to view patient's timeline
func (r *TimelineRepository) CheckDoctorAccess(doctorUserID, patientProfileID string) (bool, error) {
    // Get doctor profile ID
    var doctorProfileID string
    query := `SELECT id FROM doctor_profiles WHERE user_id = $1`
    err := r.db.Get(&doctorProfileID, query, doctorUserID)
    if err != nil {
        return false, err
    }

    // Check access permission
    var isActive bool
    query = `
        SELECT is_active
        FROM doctor_access_permissions
        WHERE doctor_id = $1 AND patient_id = $2
    `
    err = r.db.Get(&isActive, query, doctorProfileID, patientProfileID)
    if err != nil {
        return false, err
    }
    return isActive, nil
}
