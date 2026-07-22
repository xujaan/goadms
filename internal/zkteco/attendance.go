package zkteco

import (
	"bytes"
	"encoding/binary"
	"strings"
	"time"
)

// AttendanceRecord represents a single attendance record from a device.
type AttendanceRecord struct {
	UserID    string
	Timestamp time.Time
	Status    int // 0=check-in, 1=check-out, etc.
	Punch     int // punch state
}

// GetAttendances fetches all attendance records from the device.
func (c *Client) GetAttendances() ([]AttendanceRecord, error) {
	data, err := c.ExecData(CmdGetAttLog, nil)
	if err != nil {
		return nil, err
	}

	return parseAttendanceData(data)
}

func parseAttendanceData(data []byte) ([]AttendanceRecord, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Format: each record is a variable-length line.
	// Records separated by newlines or tabs.
	// Format: "user_id\ttimestamp\tstatus\tpunch\n"
	reader := bytes.NewReader(data)
	buf := make([]byte, 0, 4096)
	var records []AttendanceRecord

	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		if b == '\n' || b == '\r' {
			if len(buf) == 0 {
				continue
			}
			rec := parseAttLine(string(buf))
			if rec != nil {
				records = append(records, *rec)
			}
			buf = buf[:0]
			continue
		}
		buf = append(buf, b)
	}

	// Last line.
	if len(buf) > 0 {
		rec := parseAttLine(string(buf))
		if rec != nil {
			records = append(records, *rec)
		}
	}

	return records, nil
}

func parseAttLine(line string) *AttendanceRecord {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	fields := strings.Split(line, "\t")
	if len(fields) < 2 {
		return nil
	}

	// Parse timestamp: "2025-01-15 08:00:05"
	ts, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(fields[1]), time.Local)
	if err != nil {
		return nil
	}

	rec := &AttendanceRecord{
		UserID:    strings.TrimSpace(fields[0]),
		Timestamp: ts,
	}

	if len(fields) > 2 {
		rec.Status = atoi(strings.TrimSpace(fields[2]))
	}
	if len(fields) > 3 {
		rec.Punch = atoi(strings.TrimSpace(fields[3]))
	}

	return rec
}

func atoi(s string) int {
	var v int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + int(c-'0')
		} else {
			break
		}
	}
	return v
}

// Ensure we don't import binary unused — use a sentinel.
var _ = binary.LittleEndian
