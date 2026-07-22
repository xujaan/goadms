package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
	"github.com/jan/goadms/internal/webhook"
)

type DeviceService struct {
	deviceRepo  *repository.DeviceRepo
	rawLogRepo  *repository.RawLogRepo
	dispatcher  *webhook.Dispatcher
}

func NewDeviceService(deviceRepo *repository.DeviceRepo, rawLogRepo *repository.RawLogRepo, dispatcher *webhook.Dispatcher) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		rawLogRepo: rawLogRepo,
		dispatcher: dispatcher,
	}
}

func (s *DeviceService) GetByID(ctx context.Context, id uuid.UUID) (*model.Device, error) {
	return s.deviceRepo.GetByID(ctx, id)
}

func (s *DeviceService) GetBySN(ctx context.Context, sn string) (*model.Device, error) {
	return s.deviceRepo.GetBySN(ctx, sn)
}

func (s *DeviceService) List(ctx context.Context) ([]model.Device, error) {
	return s.deviceRepo.List(ctx, false)
}

func (s *DeviceService) Create(ctx context.Context, d *model.Device) error {
	if d.Port == 0 {
		d.Port = 4370
	}
	if d.Protocol == "" {
		d.Protocol = "zk-tcp"
	}
	if d.HandshakeConfig.Stamp == 0 {
		d.HandshakeConfig = model.DefaultHandshakeConfig()
	}
	return s.deviceRepo.Create(ctx, d)
}

func (s *DeviceService) Update(ctx context.Context, d *model.Device) error {
	return s.deviceRepo.Update(ctx, d)
}

func (s *DeviceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deviceRepo.Delete(ctx, id)
}

// RecordHandshake is called when a device performs a GET /iclock/cdata handshake.
func (s *DeviceService) RecordHandshake(ctx context.Context, sn string, ip string) error {
	wasOffline := true
	existing, _ := s.deviceRepo.GetBySN(ctx, sn)
	if existing != nil && existing.IsOnline() {
		wasOffline = false
	}

	if err := s.deviceRepo.RecordHandshake(ctx, sn, ip); err != nil {
		return err
	}

	// Get updated device for webhook.
	updated, err := s.deviceRepo.GetBySN(ctx, sn)
	if err != nil || updated == nil {
		return err
	}

	if wasOffline {
		s.dispatcher.FanOut(ctx, "device.online", updated.ID, updated.SerialNumber, map[string]any{
			"device_id":     updated.ID,
			"serial_number": updated.SerialNumber,
			"timestamp":     time.Now(),
		})
	}

	s.dispatcher.FanOut(ctx, "handshake.received", updated.ID, updated.SerialNumber, map[string]any{
		"device_sn": sn,
		"timestamp": time.Now(),
	})

	return nil
}

// ParseBodyAsTabSeparated parses POST body from ZKTeco push protocol.
// Format: tab-separated rows. Each row: employee_id, timestamp, status1..5
func ParseBodyAsTabSeparated(body string) []model.AttendanceRecord {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) == 0 {
		return nil
	}

	// Skip header if present (OPERLOG type has header row).
	// ATTLOG: "employee_id\ttimestamp\tstatus1\tstatus2\tstatus3\tstatus4\tstatus5"

	var records []model.AttendanceRecord
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			// OPERLOG type: just log count header. Skip.
			continue
		}

		// Parse timestamp in format "2025-01-15 08:00:05"
		ts, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(fields[1]), time.Local)
		if err != nil {
			continue
		}

		rec := model.AttendanceRecord{
			EmployeeID: strings.TrimSpace(fields[0]),
			Timestamp:  ts,
		}
		if len(fields) > 2 {
			rec.Status1 = parseStatusInt(fields[2])
		}
		if len(fields) > 3 {
			rec.Status2 = parseStatusInt(fields[3])
		}
		if len(fields) > 4 {
			rec.Status3 = parseStatusInt(fields[4])
		}
		if len(fields) > 5 {
			rec.Status4 = parseStatusInt(fields[5])
		}
		if len(fields) > 6 {
			rec.Status5 = parseStatusInt(fields[6])
		}

		records = append(records, rec)
	}
	return records
}

func parseStatusInt(s string) int16 {
	var v int
	for _, c := range strings.TrimSpace(s) {
		if c >= '0' && c <= '9' {
			v = v*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return int16(v)
}
