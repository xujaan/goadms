package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
	"github.com/jan/goadms/internal/webhook"
	"github.com/jan/goadms/internal/zkteco"
)

type ZkTecoService struct {
	deviceRepo     *repository.DeviceRepo
	attendanceRepo *repository.AttendanceRepo
	dispatcher     *webhook.Dispatcher
	logger         *slog.Logger
}

func NewZkTecoService(
	deviceRepo *repository.DeviceRepo,
	attendanceRepo *repository.AttendanceRepo,
	dispatcher *webhook.Dispatcher,
	logger *slog.Logger,
) *ZkTecoService {
	return &ZkTecoService{
		deviceRepo:     deviceRepo,
		attendanceRepo: attendanceRepo,
		dispatcher:     dispatcher,
		logger:         logger,
	}
}

// TestConnection tries to connect to a device via TCP.
func (s *ZkTecoService) TestConnection(ctx context.Context, deviceID uuid.UUID) error {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return fmt.Errorf("device not found")
	}

	client, err := zkteco.Connect(device.IPAddress, device.Port, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer client.Close()

	return nil
}

// PullAttendance connects to device, fetches attendance records, and stores them.
func (s *ZkTecoService) PullAttendance(ctx context.Context, deviceID uuid.UUID) (int, error) {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return 0, fmt.Errorf("device not found")
	}
	if device.IPAddress == "" {
		return 0, fmt.Errorf("device has no IP address")
	}

	start := time.Now()
	client, err := zkteco.Connect(device.IPAddress, device.Port, 30*time.Second)
	if err != nil {
		s.dispatcher.FanOut(ctx, "pull.failed", device.ID, device.SerialNumber, map[string]any{
			"device_id":     device.ID,
			"serial_number": device.SerialNumber,
			"error_message": err.Error(),
		})
		return 0, fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	records, err := client.GetAttendances()
	if err != nil {
		s.dispatcher.FanOut(ctx, "pull.failed", device.ID, device.SerialNumber, map[string]any{
			"device_id":     device.ID,
			"serial_number": device.SerialNumber,
			"error_message": err.Error(),
		})
		return 0, fmt.Errorf("get attendances: %w", err)
	}

	// Convert to model attendance records.
	var attRecords []model.AttendanceRecord
	for _, r := range records {
		attRecords = append(attRecords, model.AttendanceRecord{
			DeviceSN:   device.SerialNumber,
			EmployeeID: r.UserID,
			Timestamp:  r.Timestamp,
			Status1:    int16(r.Status),
		})
	}

	inserted, err := s.attendanceRepo.BulkInsert(ctx, attRecords, "pull")
	if err != nil {
		return 0, fmt.Errorf("store attendances: %w", err)
	}

	duration := time.Since(start)

	s.dispatcher.FanOut(ctx, "pull.completed", device.ID, device.SerialNumber, map[string]any{
		"device_id":     device.ID,
		"serial_number": device.SerialNumber,
		"records_count": inserted,
		"duration_ms":   duration.Milliseconds(),
	})

	s.logger.Info("pull attendance done",
		"device", device.Name,
		"records", inserted,
		"duration_ms", duration.Milliseconds(),
	)

	return inserted, nil
}

// GetDeviceUsers fetches user list from device via TCP.
func (s *ZkTecoService) GetDeviceUsers(ctx context.Context, deviceID uuid.UUID) ([]zkteco.UserRecord, error) {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return nil, fmt.Errorf("device not found")
	}

	client, err := zkteco.Connect(device.IPAddress, device.Port, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	return client.GetUsers()
}

// RebootDevice sends a reboot command to the device.
func (s *ZkTecoService) RebootDevice(ctx context.Context, deviceID uuid.UUID) error {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return fmt.Errorf("device not found")
	}

	client, err := zkteco.Connect(device.IPAddress, device.Port, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	return client.Reboot()
}

// SyncDeviceTime syncs the device clock to server time.
func (s *ZkTecoService) SyncDeviceTime(ctx context.Context, deviceID uuid.UUID) error {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return fmt.Errorf("device not found")
	}

	client, err := zkteco.Connect(device.IPAddress, device.Port, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	return client.SyncTime(time.Now())
}

// DeleteDeviceUser removes a user from a device by user ID via TCP.
func (s *ZkTecoService) DeleteDeviceUser(ctx context.Context, deviceID uuid.UUID, uid string) error {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return fmt.Errorf("device not found")
	}
	client, err := zkteco.Connect(device.IPAddress, device.Port, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()
	return client.DeleteUser(uid)
}

// AutoPullAll pulls attendance from all active devices.
func (s *ZkTecoService) AutoPullAll(ctx context.Context) {
	devices, err := s.deviceRepo.List(ctx, true)
	if err != nil {
		s.logger.Error("auto-pull: list devices", "error", err)
		return
	}

	for _, d := range devices {
		if d.IPAddress == "" {
			continue
		}
		n, err := s.PullAttendance(ctx, d.ID)
		if err != nil {
			s.logger.Warn("auto-pull: device failed", "device", d.Name, "error", err)
		} else {
			s.logger.Debug("auto-pull: done", "device", d.Name, "records", n)
		}
	}
}
