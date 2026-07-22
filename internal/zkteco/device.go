package zkteco

import (
	"fmt"
	"strings"
	"time"
)

// DeviceInfo represents basic device information.
type DeviceInfo struct {
	SerialNumber string `json:"serial_number"`
	Model        string `json:"model"`
	Firmware     string `json:"firmware"`
	UserCount    int    `json:"user_count"`
	AttCount     int    `json:"att_count"`
	FingerCount  int    `json:"finger_count"`
}

// GetDeviceInfo fetches device statistics via CmdGetFreeSz.
func (c *Client) GetDeviceInfo() (*DeviceInfo, error) {
	data, err := c.ExecData(CmdGetFreeSz, nil)
	if err != nil {
		return nil, err
	}

	// Response format: "user_count\tatt_count\tfinger_count\t..."
	info := &DeviceInfo{}
	if len(data) > 0 {
		parts := strings.Split(strings.TrimSpace(string(data)), "\t")
		if len(parts) > 0 {
			info.UserCount = atoi(parts[0])
		}
		if len(parts) > 1 {
			info.AttCount = atoi(parts[1])
		}
		if len(parts) > 2 {
			info.FingerCount = atoi(parts[2])
		}
	}
	return info, nil
}

// Reboot sends the restart command to the device.
func (c *Client) Reboot() error {
	return c.ExecSimple(CmdRestart, nil)
}

// SyncTime sets the device clock to the given time.
func (c *Client) SyncTime(t time.Time) error {
	payload := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
	)
	// CmdSetTime = 202 (defined in protocol.go — need to verify)
	// Some devices use a different command for time sync.
	return c.ExecSimple(202, []byte(payload))
}
