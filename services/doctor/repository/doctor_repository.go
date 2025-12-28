package repository

import (
    "database/sql"
    "health-bar/shared/models"
    "github.com/jmoiron/sqlx"
    "github.com/google/uuid"
)

type DoctorRepository struct {
    db *sqlx.DB
}

func NewDoctorRepository(db *sqlx.DB) *DoctorRepository {
    return &DoctorRepository{db: db}
}

// CreateProfile creates a doctor profile
func (r *DoctorRepository) CreateProfile(userID string, profile *models.DoctorProfile) error {
    profile.ID = uuid.New().String()
    profile.UserID = userID

    query := `
        INSERT INTO doctor_profiles (id, user_id, full_name, specialization, license_number, phone)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, user_id, full_name, specialization, license_number, phone, created_at, updated_at
    `

    return r.db.QueryRowx(query,
        profile.ID, profile.UserID, profile.FullName, profile.Specialization,
        profile.LicenseNumber, profile.Phone,
    ).StructScan(profile)
}

// GetProfileByUserID gets doctor profile by user ID
func (r *DoctorRepository) GetProfileByUserID(userID string) (*models.DoctorProfile, error) {
    profile := &models.DoctorProfile{}
    query := `
        SELECT id, user_id, full_name, specialization, license_number, phone, created_at, updated_at
        FROM doctor_profiles
        WHERE user_id = $1
    `
    err := r.db.Get(profile, query, userID)
    if err != nil {
        return nil, err
    }
    return profile, nil
}

// GetProfileByID gets doctor profile by profile ID
func (r *DoctorRepository) GetProfileByID(profileID string) (*models.DoctorProfile, error) {
    profile := &models.DoctorProfile{}
    query := `
        SELECT id, user_id, full_name, specialization, license_number, phone, created_at, updated_at
        FROM doctor_profiles
        WHERE id = $1
    `
    err := r.db.Get(profile, query, profileID)
    if err != nil {
        return nil, err
    }
    return profile, nil
}

// UpdateProfile updates doctor profile
func (r *DoctorRepository) UpdateProfile(userID string, profile *models.DoctorProfile) error {
    query := `
        UPDATE doctor_profiles
        SET full_name = $1, specialization = $2, license_number = $3, phone = $4, updated_at = NOW()
        WHERE user_id = $5
        RETURNING id, user_id, full_name, specialization, license_number, phone, created_at, updated_at
    `

    return r.db.QueryRowx(query,
        profile.FullName, profile.Specialization, profile.LicenseNumber,
        profile.Phone, userID,
    ).StructScan(profile)
}

// GetPatientProfile gets a patient profile (with permission check)
func (r *DoctorRepository) GetPatientProfile(doctorID, patientID string) (*models.PatientProfile, error) {
    // First check if doctor has access
    hasAccess, err := r.CheckAccess(doctorID, patientID)
    if err != nil {
        return nil, err
    }
    if !hasAccess {
        return nil, sql.ErrNoRows // Return not found if no access
    }

    // Get patient profile
    profile := &models.PatientProfile{}
    query := `
        SELECT id, user_id, full_name, date_of_birth, gender, phone, address, created_at, updated_at
        FROM patient_profiles
        WHERE id = $1
    `
    err = r.db.Get(profile, query, patientID)
    return profile, err
}

// CheckAccess checks if doctor has access to patient's records
func (r *DoctorRepository) CheckAccess(doctorID, patientID string) (bool, error) {
    var isActive bool
    query := `
        SELECT is_active
        FROM doctor_access_permissions
        WHERE doctor_id = $1 AND patient_id = $2
    `
    err := r.db.Get(&isActive, query, doctorID, patientID)
    if err != nil {
        if err == sql.ErrNoRows {
            return false, nil // No permission found = no access
        }
        return false, err
    }
    return isActive, nil
}

// ListAccessiblePatients lists all patients the doctor has access to
func (r *DoctorRepository) ListAccessiblePatients(doctorID string) ([]models.PatientProfile, error) {
    var patients []models.PatientProfile
    query := `
        SELECT p.id, p.user_id, p.full_name, p.date_of_birth, p.gender, p.phone, p.address, p.created_at, p.updated_at
        FROM patient_profiles p
        INNER JOIN doctor_access_permissions dap ON p.id = dap.patient_id
        WHERE dap.doctor_id = $1 AND dap.is_active = true
        ORDER BY p.full_name
    `
    err := r.db.Select(&patients, query, doctorID)
    return patients, err
}
