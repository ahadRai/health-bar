package models

import "time"

type DoctorProfile struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	FullName       string    `json:"full_name" db:"full_name"`
	Specialization string    `json:"specialization" db:"specialization"`
	LicenseNumber  string    `json:"license_number" db:"license_number"`
	Phone          string    `json:"phone" db:"phone"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
