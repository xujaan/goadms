package model

import (
	"time"

	"github.com/google/uuid"
)

type AppUser struct {
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	PasswordHash     string     `json:"-"`
	FullName         string     `json:"full_name"`
	Role             string     `json:"role"` // admin, operator
	IsActive         bool       `json:"is_active"`
	AllowedDeviceIDs []string   `json:"allowed_device_ids,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
