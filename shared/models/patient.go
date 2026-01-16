package models

import (
	"database/sql"
	"time"
)

type PatientProfile struct {
	ID          string         `json:"id" db:"id"`
	UserID      string         `json:"user_id" db:"user_id"`
	FullName    string         `json:"full_name" db:"full_name"`
	DateOfBirth time.Time      `json:"date_of_birth" db:"date_of_birth"`
	Gender      string         `json:"gender" db:"gender"`
	Phone       string         `json:"phone" db:"phone"`
	Address     string         `json:"address" db:"address"`
	ShareToken  sql.NullString `json:"share_token" db:"share_token"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}
