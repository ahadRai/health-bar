package repository

import (
	"health-bar/shared/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

// GenerateShareToken generates a new share token for the profile
func (r *PatientRepository) GenerateShareToken(profileID string) (string, error) {
	token := uuid.New().String()
	query := `
        UPDATE patient_profiles
        SET share_token = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := r.db.Exec(query, token, profileID)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetProfileByShareToken gets patient profile by share token
func (r *PatientRepository) GetProfileByShareToken(token string) (*models.PatientProfile, error) {
	profile := &models.PatientProfile{}
	query := `
        SELECT id, user_id, full_name, date_of_birth, gender, phone, address, share_token, created_at, updated_at
        FROM patient_profiles
        WHERE share_token = $1
    `
	err := r.db.Get(profile, query, token)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
