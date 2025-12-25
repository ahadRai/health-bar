package models

import "time"

type Prescription struct {
	ID         string    `json:"id" db:"id"`
	PatientID  string    `json:"patient_id" db:"patient_id"`
	FileName   string    `json:"file_name" db:"file_name"`
	FileType   string    `json:"file_type" db:"file_type"`
	FileSize   int64     `json:"file_size" db:"file_size"`
	FilePath   string    `json:"file_path" db:"file_path"`
	UploadDate time.Time `json:"upload_date" db:"upload_date"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
