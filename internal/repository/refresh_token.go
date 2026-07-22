package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepo struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	hash := tokenHash(token)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (app_user_id, token_hash, expires_at) VALUES ($1,$2,$3)`,
		userID, hash, expiresAt)
	if err != nil {
		return fmt.Errorf("create refresh_token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepo) ValidateAndDelete(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	hash := tokenHash(token)
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE app_user_id = $1 AND token_hash = $2 AND expires_at > NOW()`,
		userID, hash)
	if err != nil {
		return false, fmt.Errorf("validate refresh_token: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *RefreshTokenRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE app_user_id = $1`, userID)
	return err
}

func (r *RefreshTokenRepo) CleanupExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
