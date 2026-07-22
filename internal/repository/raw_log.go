package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jan/goadms/internal/model"
)

type RawLogRepo struct {
	pool *pgxpool.Pool
}

func NewRawLogRepo(pool *pgxpool.Pool) *RawLogRepo {
	return &RawLogRepo{pool: pool}
}

func (r *RawLogRepo) Insert(ctx context.Context, log *model.RawLog) error {
	log.CreatedAt = time.Now()
	qp, _ := json.Marshal(log.QueryParams)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO raw_logs (device_sn, request_method, request_uri, query_params, request_body, response_body, log_type, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		log.DeviceSN, log.RequestMethod, log.RequestURI, qp, log.RequestBody, log.ResponseBody, log.LogType, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert raw_log: %w", err)
	}
	return nil
}

func (r *RawLogRepo) List(ctx context.Context, sn string, limit int) ([]model.RawLog, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT id, device_sn, request_method, request_uri, query_params, request_body, response_body, log_type, created_at
	      FROM raw_logs`
	args := []any{}
	if sn != "" {
		q += ` WHERE device_sn = $1`
		args = append(args, sn)
	}
	q += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list raw_logs: %w", err)
	}
	defer rows.Close()

	var logs []model.RawLog
	for rows.Next() {
		var l model.RawLog
		var qp []byte
		if err := rows.Scan(&l.ID, &l.DeviceSN, &l.RequestMethod, &l.RequestURI, &qp, &l.RequestBody, &l.ResponseBody, &l.LogType, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan raw_log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
