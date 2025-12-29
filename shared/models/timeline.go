package models

import "time"

type HospitalVisit struct {
    ID           string    `json:"id" db:"id"`
    PatientID    string    `json:"patient_id" db:"patient_id"`
    HospitalName string    `json:"hospital_name" db:"hospital_name"`
    VisitDate    time.Time `json:"visit_date" db:"visit_date"`
    Reason       string    `json:"reason" db:"reason"`
    Notes        string    `json:"notes" db:"notes"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
