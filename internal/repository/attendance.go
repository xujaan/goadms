package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jan/goadms/internal/model"
)

type AttendanceRepo struct {
	pool *pgxpool.Pool
}

func NewAttendanceRepo(pool *pgxpool.Pool) *AttendanceRepo {
	return &AttendanceRepo{pool: pool}
}

func (r *AttendanceRepo) BulkInsert(ctx context.Context, records []model.AttendanceRecord, source string) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	// Use COPY protocol via pgx for bulk insert efficiency.
	// Fallback to batch INSERT ... ON CONFLICT DO NOTHING.
	now := time.Now()
	inserted := 0

	batch := &pgx.Batch{}
	for _, rec := range records {
		raw, _ := json.Marshal(rec)
		batch.Queue(
			`INSERT INTO attendances (device_sn, employee_id, timestamp, status1, status2, status3, status4, status5, source, raw_payload, created_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			 ON CONFLICT (device_sn, employee_id, timestamp, COALESCE(status1, -1), COALESCE(status2, -1), COALESCE(status3, -1), COALESCE(status4, -1), COALESCE(status5, -1)) DO NOTHING`,
			rec.DeviceSN, rec.EmployeeID, rec.Timestamp,
			rec.Status1, rec.Status2, rec.Status3, rec.Status4, rec.Status5,
			source, string(raw), now,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range records {
		_, err := br.Exec()
		if err == nil {
			inserted++
		}
	}

	return inserted, br.Close()
}

func (r *AttendanceRepo) List(ctx context.Context, filter AttendanceFilter) ([]model.Attendance, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	argN := 1

	if filter.DeviceSN != "" {
		where += fmt.Sprintf(" AND device_sn = $%d", argN)
		args = append(args, filter.DeviceSN)
		argN++
	}
	if filter.EmployeeID != "" {
		where += fmt.Sprintf(" AND employee_id = $%d", argN)
		args = append(args, filter.EmployeeID)
		argN++
	}
	if filter.Source != "" {
		where += fmt.Sprintf(" AND source = $%d", argN)
		args = append(args, filter.Source)
		argN++
	}
	if !filter.DateFrom.IsZero() {
		where += fmt.Sprintf(" AND timestamp >= $%d", argN)
		args = append(args, filter.DateFrom)
		argN++
	}
	if !filter.DateTo.IsZero() {
		where += fmt.Sprintf(" AND timestamp <= $%d", argN)
		args = append(args, filter.DateTo)
		argN++
	}

	// Count
	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM attendances %s", where)
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count attendances: %w", err)
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	q := fmt.Sprintf(
		`SELECT id, device_id, device_sn, employee_id, timestamp, status1, status2, status3, status4, status5, source, raw_payload, created_at
		 FROM attendances %s ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`,
		where, argN, argN+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list attendances: %w", err)
	}
	defer rows.Close()

	var records []model.Attendance
	for rows.Next() {
		var a model.Attendance
		if err := rows.Scan(&a.ID, &a.DeviceID, &a.DeviceSN, &a.EmployeeID, &a.Timestamp,
			&a.Status1, &a.Status2, &a.Status3, &a.Status4, &a.Status5,
			&a.Source, &a.RawPayload, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan attendance: %w", err)
		}
		records = append(records, a)
	}
	return records, total, rows.Err()
}

func (r *AttendanceRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM attendances WHERE id = $1`, id)
	return err
}

type AttendanceFilter struct {
	DeviceSN   string
	EmployeeID string
	Source     string
	DateFrom   time.Time
	DateTo     time.Time
	Limit      int
	Offset     int
}
