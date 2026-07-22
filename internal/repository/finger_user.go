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

type FingerUserRepo struct {
	pool *pgxpool.Pool
}

func NewFingerUserRepo(pool *pgxpool.Pool) *FingerUserRepo {
	return &FingerUserRepo{pool: pool}
}

func (r *FingerUserRepo) List(ctx context.Context) ([]model.FingerprintUser, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, employee_code, full_name, department, is_active, created_at, updated_at
		 FROM fingerprint_users ORDER BY full_name`)
	if err != nil {
		return nil, fmt.Errorf("list fingerprint_users: %w", err)
	}
	defer rows.Close()

	var users []model.FingerprintUser
	for rows.Next() {
		var u model.FingerprintUser
		if err := rows.Scan(&u.ID, &u.EmployeeCode, &u.FullName, &u.Department, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan fingerprint_user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *FingerUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.FingerprintUser, error) {
	var u model.FingerprintUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, employee_code, full_name, department, is_active, created_at, updated_at
		 FROM fingerprint_users WHERE id = $1`, id,
	).Scan(&u.ID, &u.EmployeeCode, &u.FullName, &u.Department, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get fingerprint_user: %w", err)
	}
	return &u, nil
}

func (r *FingerUserRepo) Create(ctx context.Context, u *model.FingerprintUser) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO fingerprint_users (id, employee_code, full_name, department, is_active, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		u.ID, u.EmployeeCode, u.FullName, u.Department, u.IsActive, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create fingerprint_user: %w", err)
	}
	return nil
}

func (r *FingerUserRepo) Update(ctx context.Context, u *model.FingerprintUser) error {
	u.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE fingerprint_users SET employee_code=$1, full_name=$2, department=$3, is_active=$4, updated_at=$5 WHERE id=$6`,
		u.EmployeeCode, u.FullName, u.Department, u.IsActive, u.UpdatedAt, u.ID)
	return err
}

func (r *FingerUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM fingerprint_users WHERE id = $1`, id)
	return err
}
