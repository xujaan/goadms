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

type AppUserRepo struct {
	pool *pgxpool.Pool
}

func NewAppUserRepo(pool *pgxpool.Pool) *AppUserRepo {
	return &AppUserRepo{pool: pool}
}

func (r *AppUserRepo) GetByUsername(ctx context.Context, username string) (*model.AppUser, error) {
	var u model.AppUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, full_name, role, is_active, allowed_device_ids, created_at, updated_at
		 FROM app_users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.AllowedDeviceIDs, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query app_user: %w", err)
	}
	return &u, nil
}

func (r *AppUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.AppUser, error) {
	var u model.AppUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, full_name, role, is_active, allowed_device_ids, created_at, updated_at
		 FROM app_users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.AllowedDeviceIDs, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query app_user: %w", err)
	}
	return &u, nil
}

func (r *AppUserRepo) List(ctx context.Context) ([]model.AppUser, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, password_hash, full_name, role, is_active, allowed_device_ids, created_at, updated_at
		 FROM app_users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list app_users: %w", err)
	}
	defer rows.Close()

	var users []model.AppUser
	for rows.Next() {
		var u model.AppUser
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.AllowedDeviceIDs, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan app_user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *AppUserRepo) Create(ctx context.Context, u *model.AppUser) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO app_users (id, username, password_hash, full_name, role, is_active, allowed_device_ids, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		u.ID, u.Username, u.PasswordHash, u.FullName, u.Role, u.IsActive, u.AllowedDeviceIDs, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create app_user: %w", err)
	}
	return nil
}

func (r *AppUserRepo) Update(ctx context.Context, u *model.AppUser) error {
	u.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE app_users SET full_name=$1, role=$2, is_active=$3, allowed_device_ids=$4, updated_at=$5 WHERE id=$6`,
		u.FullName, u.Role, u.IsActive, u.AllowedDeviceIDs, u.UpdatedAt, u.ID)
	if err != nil {
		return fmt.Errorf("update app_user: %w", err)
	}
	return nil
}

func (r *AppUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM app_users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete app_user: %w", err)
	}
	return nil
}
