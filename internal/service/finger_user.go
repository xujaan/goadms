package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
	"github.com/jan/goadms/internal/zkteco"
	"log/slog"
)

type FingerUserService struct {
	fingerRepo *repository.FingerUserRepo
	deviceRepo *repository.DeviceRepo
	logger     *slog.Logger
}

func NewFingerUserService(fingerRepo *repository.FingerUserRepo, deviceRepo *repository.DeviceRepo, logger *slog.Logger) *FingerUserService {
	return &FingerUserService{fingerRepo: fingerRepo, deviceRepo: deviceRepo, logger: logger}
}

func (s *FingerUserService) List(ctx context.Context) ([]model.FingerprintUser, error) {
	return s.fingerRepo.List(ctx)
}

func (s *FingerUserService) Create(ctx context.Context, u *model.FingerprintUser) error {
	return s.fingerRepo.Create(ctx, u)
}

func (s *FingerUserService) Update(ctx context.Context, u *model.FingerprintUser) error {
	return s.fingerRepo.Update(ctx, u)
}

func (s *FingerUserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.fingerRepo.Delete(ctx, id)
}

// SyncToDevice pushes all fingerprint users to a device via TCP.
func (s *FingerUserService) SyncToDevice(ctx context.Context, userID, deviceID uuid.UUID) error {
	user, err := s.fingerRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}

	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return fmt.Errorf("device not found")
	}

	client, err := zkteco.Connect(device.IPAddress, device.Port, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	return client.SetUser(zkteco.UserRecord{
		UserID:    user.EmployeeCode,
		Name:      user.FullName,
		Privilege: 0,
	})
}

// SyncAllToDevice pushes all fingerprint users to a device.
func (s *FingerUserService) SyncAllToDevice(ctx context.Context, deviceID uuid.UUID) (int, error) {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil || device == nil {
		return 0, fmt.Errorf("device not found")
	}

	users, err := s.fingerRepo.List(ctx)
	if err != nil {
		return 0, err
	}

	client, err := zkteco.Connect(device.IPAddress, device.Port, 30*time.Second)
	if err != nil {
		return 0, fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	synced := 0
	for _, u := range users {
		if err := client.SetUser(zkteco.UserRecord{
			UserID:    u.EmployeeCode,
			Name:      u.FullName,
			Privilege: 0,
		}); err != nil {
			s.logger.Warn("sync user to device failed", "user", u.FullName, "device", device.Name, "error", err)
		} else {
			synced++
		}
	}

	return synced, nil
}
