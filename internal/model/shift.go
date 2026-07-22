package model

import (
	"time"

	"github.com/google/uuid"
)

type Shift struct {
	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	StartTime            string    `json:"start_time"` // HH:MM
	EndTime              string    `json:"end_time"`   // HH:MM
	BreakMinutes         int       `json:"break_minutes"`
	LateToleranceMinutes int       `json:"late_tolerance_minutes"`
	OvertimeAfterMinutes int       `json:"overtime_after_minutes"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type UserShift struct {
	ID                int64     `json:"id"`
	FingerprintUserID uuid.UUID `json:"fingerprint_user_id"`
	ShiftID           uuid.UUID `json:"shift_id"`
	EffectiveDate     string    `json:"effective_date"` // YYYY-MM-DD
}
