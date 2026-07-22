package model

import (
	"time"

	"github.com/google/uuid"
)

type FingerprintUser struct {
	ID           uuid.UUID `json:"id"`
	EmployeeCode string    `json:"employee_code"`
	FullName     string    `json:"full_name"`
	Department   string    `json:"department,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DeviceUser struct {
	ID                int64      `json:"id"`
	DeviceID          uuid.UUID  `json:"device_id"`
	FingerprintUserID *uuid.UUID `json:"fingerprint_user_id,omitempty"`
	UID               int        `json:"uid"`
	EmployeeCode      string     `json:"employee_code,omitempty"`
	FullName          string     `json:"full_name,omitempty"`
	Privilege         int        `json:"privilege"`
	FingerprintCount  int        `json:"fingerprint_count"`
	SyncedAt          *time.Time `json:"synced_at,omitempty"`
}
