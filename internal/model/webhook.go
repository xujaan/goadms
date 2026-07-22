package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type WebhookConfig struct {
	ID        uuid.UUID  `json:"id"`
	DeviceID  uuid.UUID  `json:"device_id"`
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	Secret    string     `json:"secret,omitempty"`
	Events    []string   `json:"events"`
	Headers   map[string]string `json:"headers,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type WebhookDelivery struct {
	ID              int64      `json:"id"`
	WebhookConfigID uuid.UUID  `json:"webhook_config_id"`
	EventType       string     `json:"event_type"`
	Payload         []byte     `json:"payload"`
	RequestURL      string     `json:"request_url"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	ResponseStatus  int        `json:"response_status,omitempty"`
	ResponseBody    string     `json:"response_body,omitempty"`
	Status          string     `json:"status"` // pending, success, failed, retrying
	AttemptCount    int        `json:"attempt_count"`
	MaxAttempts     int        `json:"max_attempts"`
	NextAttemptAt   *time.Time `json:"next_attempt_at,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	DurationMs      int        `json:"duration_ms,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// WebhookEvent is the envelope for webhook dispatch.
type WebhookEvent struct {
	Event     string          `json:"event"`
	DeviceID  uuid.UUID       `json:"device_id"`
	DeviceSN  string          `json:"device_sn"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}
