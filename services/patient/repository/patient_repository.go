package repository

import (
    "health-bar/shared/models"
    "github.com/jmoiron/sqlx"
    "github.com/google/uuid"
)

type PatientRepository struct {
    db *sqlx.DB
}

func NewPatientRepository(db *sqlx.DB) *PatientRepository {
    return &PatientRepository{db: db}
}

// CreateProfile creates a patient profile
func (r *PatientRepository) CreateProfile(userID string, profile *models.PatientProfile) error {
    profile.ID = uuid.New().String()
    profile.UserID = userID

    query := `
        INSERT INTO patient_profiles (id, user_id, full_name, date_of_birth, gender, phone, address)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, user_id, full_name, date_of_birth, gender, phone, address, created_at, updated_at
    `

    return r.db.QueryRowx(query,
        profile.ID, profile.UserID, profile.FullName, profile.DateOfBirth,
        profile.Gender, profile.Phone, profile.Address,
    ).StructScan(profile)
}

// GetProfileByUserID gets patient profile by user ID
func (r *PatientRepository) GetProfileByUserID(userID string) (*models.PatientProfile, error) {
    profile := &models.PatientProfile{}
    query := `
        SELECT id, user_id, full_name, date_of_birth, gender, phone, address, created_at, updated_at
        FROM patient_profiles
        WHERE user_id = $1
    `
    err := r.db.Get(profile, query, userID)
    if err != nil {
        return nil, err
    }
    return profile, nil
}

// GetProfileByID gets patient profile by profile ID
func (r *PatientRepository) GetProfileByID(profileID string) (*models.PatientProfile, error) {
    profile := &models.PatientProfile{}
    query := `
        SELECT id, user_id, full_name, date_of_birth, gender, phone, address, created_at, updated_at
        FROM patient_profiles
        WHERE id = $1
    `
    err := r.db.Get(profile, query, profileID)
    if err != nil {
        return nil, err
    }
    return profile, nil
}

// UpdateProfile updates patient profile
func (r *PatientRepository) UpdateProfile(userID string, profile *models.PatientProfile) error {
    query := `
        UPDATE patient_profiles
        SET full_name = $1, date_of_birth = $2, gender = $3, phone = $4, address = $5, updated_at = NOW()
        WHERE user_id = $6
        RETURNING id, user_id, full_name, date_of_birth, gender, phone, address, created_at, updated_at
    `

    return r.db.QueryRowx(query,
        profile.FullName, profile.DateOfBirth, profile.Gender,
        profile.Phone, profile.Address, userID,
    ).StructScan(profile)
}

// GrantAccess grants a doctor access to patient's records
func (r *PatientRepository) GrantAccess(patientID, doctorID string) error {
    permission := &models.DoctorAccessPermission{
        ID:        uuid.New().String(),
        PatientID: patientID,
        DoctorID:  doctorID,
        IsActive:  true,
    }

    query := `
        INSERT INTO doctor_access_permissions (id, patient_id, doctor_id, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (patient_id, doctor_id) 
        DO UPDATE SET is_active = true, revoked_at = NULL, granted_at = NOW()
    `

    _, err := r.db.Exec(query, permission.ID, permission.PatientID, permission.DoctorID, permission.IsActive)
    return err
}

// RevokeAccess revokes a doctor's access to patient's records
func (r *PatientRepository) RevokeAccess(patientID, doctorID string) error {
    query := `
        UPDATE doctor_access_permissions
        SET is_active = false, revoked_at = NOW()
        WHERE patient_id = $1 AND doctor_id = $2
    `

    _, err := r.db.Exec(query, patientID, doctorID)
    return err
}

// ListPermissions lists all doctors who have access to patient's records
func (r *PatientRepository) ListPermissions(patientID string) ([]models.DoctorAccessPermission, error) {
    var permissions []models.DoctorAccessPermission
    query := `
        SELECT id, patient_id, doctor_id, granted_at, revoked_at, is_active
        FROM doctor_access_permissions
        WHERE patient_id = $1
        ORDER BY granted_at DESC
    `

    err := r.db.Select(&permissions, query, patientID)
    return permissions, err
}

// CheckAccess checks if a doctor has access to a patient's records
func (r *PatientRepository) CheckAccess(patientID, doctorID string) (bool, error) {
    var isActive bool
    query := `
        SELECT is_active
        FROM doctor_access_permissions
        WHERE patient_id = $1 AND doctor_id = $2
    `

    err := r.db.Get(&isActive, query, patientID, doctorID)
    if err != nil {
        return false, err
    }
    return isActive, nil
}
