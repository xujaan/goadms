package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
)

type ShiftService struct {
	shiftRepo *repository.ShiftRepo
}

func NewShiftService(shiftRepo *repository.ShiftRepo) *ShiftService {
	return &ShiftService{shiftRepo: shiftRepo}
}

func (s *ShiftService) List(ctx context.Context) ([]model.Shift, error) {
	return s.shiftRepo.List(ctx)
}

func (s *ShiftService) GetByID(ctx context.Context, id uuid.UUID) (*model.Shift, error) {
	return s.shiftRepo.GetByID(ctx, id)
}

func (s *ShiftService) Create(ctx context.Context, shift *model.Shift) error {
	return s.shiftRepo.Create(ctx, shift)
}

func (s *ShiftService) Update(ctx context.Context, shift *model.Shift) error {
	return s.shiftRepo.Update(ctx, shift)
}

func (s *ShiftService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.shiftRepo.Delete(ctx, id)
}

func (s *ShiftService) AssignUser(ctx context.Context, shiftID, userID uuid.UUID, date string) error {
	return s.shiftRepo.AssignUser(ctx, shiftID, userID, date)
}

func (s *ShiftService) UnassignUser(ctx context.Context, userID, shiftID uuid.UUID, date string) error {
	return s.shiftRepo.UnassignUser(ctx, userID, shiftID, date)
}

func (s *ShiftService) GetAssignedUsers(ctx context.Context, shiftID uuid.UUID) ([]model.UserShift, error) {
	return s.shiftRepo.GetAssignedUsers(ctx, shiftID)
}
