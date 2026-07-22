package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jan/goadms/internal/model"
)

type ShiftRepo struct {
	pool *pgxpool.Pool
}

func NewShiftRepo(pool *pgxpool.Pool) *ShiftRepo {
	return &ShiftRepo{pool: pool}
}

func (r *ShiftRepo) List(ctx context.Context) ([]model.Shift, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, start_time, end_time, break_minutes, late_tolerance_minutes, overtime_after_minutes, is_active, created_at, updated_at
		 FROM shifts ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list shifts: %w", err)
	}
	defer rows.Close()

	var shifts []model.Shift
	for rows.Next() {
		var s model.Shift
		var startTime, endTime time.Time // TIME is returned as time.Time from pgx
		if err := rows.Scan(&s.ID, &s.Name, &startTime, &endTime,
			&s.BreakMinutes, &s.LateToleranceMinutes, &s.OvertimeAfterMinutes,
			&s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan shift: %w", err)
		}
		s.StartTime = startTime.Format("15:04")
		s.EndTime = endTime.Format("15:04")
		shifts = append(shifts, s)
	}
	return shifts, rows.Err()
}

func (r *ShiftRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Shift, error) {
	var s model.Shift
	var startTime, endTime time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, start_time, end_time, break_minutes, late_tolerance_minutes, overtime_after_minutes, is_active, created_at, updated_at
		 FROM shifts WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &startTime, &endTime,
		&s.BreakMinutes, &s.LateToleranceMinutes, &s.OvertimeAfterMinutes,
		&s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get shift: %w", err)
	}
	s.StartTime = startTime.Format("15:04")
	s.EndTime = endTime.Format("15:04")
	return &s, nil
}

func (r *ShiftRepo) Create(ctx context.Context, s *model.Shift) error {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO shifts (id, name, start_time, end_time, break_minutes, late_tolerance_minutes, overtime_after_minutes, is_active, created_at, updated_at)
		 VALUES ($1,$2,$3::TIME,$4::TIME,$5,$6,$7,$8,$9,$10)`,
		s.ID, s.Name, s.StartTime, s.EndTime, s.BreakMinutes, s.LateToleranceMinutes, s.OvertimeAfterMinutes, s.IsActive, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create shift: %w", err)
	}
	return nil
}

func (r *ShiftRepo) Update(ctx context.Context, s *model.Shift) error {
	s.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE shifts SET name=$1, start_time=$2::TIME, end_time=$3::TIME, break_minutes=$4, late_tolerance_minutes=$5, overtime_after_minutes=$6, is_active=$7, updated_at=$8 WHERE id=$9`,
		s.Name, s.StartTime, s.EndTime, s.BreakMinutes, s.LateToleranceMinutes, s.OvertimeAfterMinutes, s.IsActive, s.UpdatedAt, s.ID)
	if err != nil {
		return fmt.Errorf("update shift: %w", err)
	}
	return nil
}

func (r *ShiftRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM shifts WHERE id = $1`, id)
	return err
}

func (r *ShiftRepo) AssignUser(ctx context.Context, shiftID, fingerprintUserID uuid.UUID, effectiveDate string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_shifts (fingerprint_user_id, shift_id, effective_date) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		fingerprintUserID, shiftID, effectiveDate)
	return err
}

func (r *ShiftRepo) UnassignUser(ctx context.Context, fingerprintUserID, shiftID uuid.UUID, effectiveDate string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_shifts WHERE fingerprint_user_id=$1 AND shift_id=$2 AND effective_date=$3`,
		fingerprintUserID, shiftID, effectiveDate)
	return err
}

func (r *ShiftRepo) GetAssignedUsers(ctx context.Context, shiftID uuid.UUID) ([]model.UserShift, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, fingerprint_user_id, shift_id, effective_date::TEXT FROM user_shifts WHERE shift_id = $1 ORDER BY effective_date`, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uShifts []model.UserShift
	for rows.Next() {
		var us model.UserShift
		if err := rows.Scan(&us.ID, &us.FingerprintUserID, &us.ShiftID, &us.EffectiveDate); err != nil {
			return nil, err
		}
		uShifts = append(uShifts, us)
	}
	return uShifts, rows.Err()
}
