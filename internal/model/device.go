package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// HandshakeConfig is stored as JSONB in the devices table.
// Mirrors the ZKTeco push-protocol options response.
type HandshakeConfig struct {
	Stamp         int    `json:"stamp"`
	OpStamp       int    `json:"op_stamp"`
	ErrorDelay    int    `json:"error_delay"`
	Delay         int    `json:"delay"`
	ResLogDay     int    `json:"res_log_day"`
	ResLogCount   int    `json:"res_log_del_count"`
	ResLogDelCount int   `json:"res_log_count"`
	TransTimes    string `json:"trans_times"`
	TransInterval int    `json:"trans_interval"`
	TransFlag     string `json:"trans_flag"`
	TimeZone      int    `json:"time_zone"`
	Realtime      int    `json:"realtime"`
	Encrypt       int    `json:"encrypt"`
}

func DefaultHandshakeConfig() HandshakeConfig {
	return HandshakeConfig{
		Stamp:          9999,
		OpStamp:        9999,
		ErrorDelay:     60,
		Delay:          30,
		ResLogDay:      18250,
		ResLogDelCount: 10000,
		ResLogCount:    50000,
		TransTimes:     "00:00;14:05",
		TransInterval:  1,
		TransFlag:      "1111000000",
		TimeZone:       7,
		Realtime:       1,
		Encrypt:        0,
	}
}

type Device struct {
	ID               uuid.UUID        `json:"id"`
	Name             string           `json:"name"`
	SerialNumber     string           `json:"serial_number"`
	IPAddress        string           `json:"ip_address,omitempty"`
	Port             int              `json:"port"`
	Location         string           `json:"location,omitempty"`
	Brand            string           `json:"brand,omitempty"`
	Protocol         string           `json:"protocol"` // zk-tcp, adms-http
	Timezone         string           `json:"timezone"`
	HandshakeConfig  HandshakeConfig  `json:"handshake_config"`
	LastHandshakeAt  *time.Time       `json:"last_handshake_at,omitempty"`
	IsActive         bool             `json:"is_active"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// ScanHandshakeConfig unmarshals JSONB from database.
func (d *Device) ScanHandshakeConfig(src any) error {
	if src == nil {
		d.HandshakeConfig = DefaultHandshakeConfig()
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		d.HandshakeConfig = DefaultHandshakeConfig()
		return nil
	}
	return json.Unmarshal(b, &d.HandshakeConfig)
}

// IsOnline returns true if last handshake was within 5 minutes.
func (d *Device) IsOnline() bool {
	if d.LastHandshakeAt == nil {
		return false
	}
	return time.Since(*d.LastHandshakeAt) < 5*time.Minute
}
