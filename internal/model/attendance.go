package model

import (
	"time"

	"github.com/google/uuid"
)

type Attendance struct {
	ID          int64      `json:"id"`
	DeviceID    *uuid.UUID `json:"device_id,omitempty"`
	DeviceSN    string     `json:"device_sn"`
	EmployeeID  string     `json:"employee_id"`
	Timestamp   time.Time  `json:"timestamp"`
	Status1     int16      `json:"status1"`
	Status2     int16      `json:"status2"`
	Status3     int16      `json:"status3"`
	Status4     int16      `json:"status4"`
	Status5     int16      `json:"status5"`
	Source      string     `json:"source"` // push, pull
	RawPayload  string     `json:"raw_payload,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AttendanceRecord is a simplified input from ZKTeco push/TCP pull.
type AttendanceRecord struct {
	DeviceSN   string    `json:"device_sn"`
	EmployeeID string    `json:"employee_id"`
	Timestamp  time.Time `json:"timestamp"`
	Status1    int16     `json:"status1"`
	Status2    int16     `json:"status2"`
	Status3    int16     `json:"status3"`
	Status4    int16     `json:"status4"`
	Status5    int16     `json:"status5"`
}
