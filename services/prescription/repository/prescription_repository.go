package repository

import (
	"health-bar/shared/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PrescriptionRepository struct {
	db *sqlx.DB
}

func NewPrescriptionRepository(db *sqlx.DB) *PrescriptionRepository {
	return &PrescriptionRepository{db: db}
}

// CreatePrescription creates a new prescription record
func (r *PrescriptionRepository) CreatePrescription(patientID string, prescription *models.Prescription) error {
	prescription.ID = uuid.New().String()
	prescription.PatientID = patientID

	query := `
        INSERT INTO prescriptions (id, patient_id, visit_id, file_name, file_type, file_size, file_path, upload_date)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
        RETURNING id, patient_id, visit_id, file_name, file_type, file_size, file_path, upload_date, created_at
    `

	return r.db.QueryRowx(query,
		prescription.ID, prescription.PatientID, prescription.VisitID, prescription.FileName,
		prescription.FileType, prescription.FileSize, prescription.FilePath,
	).StructScan(prescription)
}

// GetPrescriptionByID gets a prescription by ID
func (r *PrescriptionRepository) GetPrescriptionByID(prescriptionID string) (*models.Prescription, error) {
	prescription := &models.Prescription{}
	query := `
        SELECT id, patient_id, visit_id, file_name, file_type, file_size, file_path, upload_date, created_at
        FROM prescriptions
        WHERE id = $1
    `
	err := r.db.Get(prescription, query, prescriptionID)
	return prescription, err
}

// GetPrescriptionsByPatientID gets all prescriptions for a patient
func (r *PrescriptionRepository) GetPrescriptionsByPatientID(patientID string) ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	query := `
        SELECT id, patient_id, visit_id, file_name, file_type, file_size, file_path, upload_date, created_at
        FROM prescriptions
        WHERE patient_id = $1
        ORDER BY upload_date DESC
    `
	err := r.db.Select(&prescriptions, query, patientID)
	return prescriptions, err
}

// DeletePrescription deletes a prescription
func (r *PrescriptionRepository) DeletePrescription(prescriptionID string) error {
	query := `DELETE FROM prescriptions WHERE id = $1`
	_, err := r.db.Exec(query, prescriptionID)
	return err
}

// GetPatientIDByPrescriptionID gets the patient ID for a prescription
func (r *PrescriptionRepository) GetPatientIDByPrescriptionID(prescriptionID string) (string, error) {
	var patientID string
	query := `SELECT patient_id FROM prescriptions WHERE id = $1`
	err := r.db.Get(&patientID, query, prescriptionID)
	return patientID, err
}

// GetPatientProfileIDByUserID gets patient profile ID from user ID
func (r *PrescriptionRepository) GetPatientProfileIDByUserID(userID string) (string, error) {
	var profileID string
	query := `SELECT id FROM patient_profiles WHERE user_id = $1`
	err := r.db.Get(&profileID, query, userID)
	return profileID, err
}
