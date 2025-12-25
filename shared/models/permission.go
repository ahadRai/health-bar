package models

import "time"

type DoctorAccessPermission struct {
	ID        string     `json:"id" db:"id"`
	PatientID string     `json:"patient_id" db:"patient_id"`
	DoctorID  string     `json:"doctor_id" db:"doctor_id"`
	GrantedAt time.Time  `json:"granted_at" db:"granted_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	IsActive  bool       `json:"is_active" db:"is_active"`
}
