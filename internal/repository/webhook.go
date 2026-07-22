package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jan/goadms/internal/model"
)

type WebhookRepo struct {
	pool *pgxpool.Pool
}

func NewWebhookRepo(pool *pgxpool.Pool) *WebhookRepo {
	return &WebhookRepo{pool: pool}
}

// --- WebhookConfig ---

func (r *WebhookRepo) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]model.WebhookConfig, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, device_id, name, url, COALESCE(secret,''), events, headers, is_active, created_at, updated_at
		 FROM webhook_configs WHERE device_id = $1 ORDER BY name`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list webhook_configs: %w", err)
	}
	defer rows.Close()

	var configs []model.WebhookConfig
	for rows.Next() {
		c, err := scanWebhookConfig(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, *c)
	}
	return configs, rows.Err()
}

func (r *WebhookRepo) ListActiveByDeviceAndEvent(ctx context.Context, deviceID uuid.UUID, eventType string) ([]model.WebhookConfig, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, device_id, name, url, COALESCE(secret,''), events, headers, is_active, created_at, updated_at
		 FROM webhook_configs WHERE device_id = $1 AND is_active = true AND $2 = ANY(events)`, deviceID, eventType)
	if err != nil {
		return nil, fmt.Errorf("list active webhooks: %w", err)
	}
	defer rows.Close()

	var configs []model.WebhookConfig
	for rows.Next() {
		c, err := scanWebhookConfig(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, *c)
	}
	return configs, rows.Err()
}

func (r *WebhookRepo) GetAllActiveByEvent(ctx context.Context, eventType string) ([]model.WebhookConfig, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, device_id, name, url, COALESCE(secret,''), events, headers, is_active, created_at, updated_at
		 FROM webhook_configs WHERE is_active = true AND $1 = ANY(events)`, eventType)
	if err != nil {
		return nil, fmt.Errorf("list all active webhooks: %w", err)
	}
	defer rows.Close()

	var configs []model.WebhookConfig
	for rows.Next() {
		c, err := scanWebhookConfig(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, *c)
	}
	return configs, rows.Err()
}

func (r *WebhookRepo) GetConfigByID(ctx context.Context, id uuid.UUID) (*model.WebhookConfig, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, device_id, name, url, COALESCE(secret,''), events, headers, is_active, created_at, updated_at
		 FROM webhook_configs WHERE id = $1`, id)
	c, err := scanWebhookConfig(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *WebhookRepo) CreateConfig(ctx context.Context, c *model.WebhookConfig) error {
	c.ID = uuid.New()
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	if c.Events == nil {
		c.Events = []string{}
	}
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	hdrs, _ := json.Marshal(c.Headers)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhook_configs (id, device_id, name, url, secret, events, headers, is_active, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.DeviceID, c.Name, c.URL, nullStr(c.Secret), c.Events, hdrs, c.IsActive, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create webhook_config: %w", err)
	}
	return nil
}

func (r *WebhookRepo) UpdateConfig(ctx context.Context, c *model.WebhookConfig) error {
	c.UpdatedAt = time.Now()
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	hdrs, _ := json.Marshal(c.Headers)
	_, err := r.pool.Exec(ctx,
		`UPDATE webhook_configs SET name=$1, url=$2, secret=$3, events=$4, headers=$5, is_active=$6, updated_at=$7
		 WHERE id=$8`,
		c.Name, c.URL, nullStr(c.Secret), c.Events, hdrs, c.IsActive, c.UpdatedAt, c.ID)
	if err != nil {
		return fmt.Errorf("update webhook_config: %w", err)
	}
	return nil
}

func (r *WebhookRepo) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM webhook_configs WHERE id = $1`, id)
	return err
}

// --- WebhookDelivery ---

func (r *WebhookRepo) InsertDelivery(ctx context.Context, d *model.WebhookDelivery) error {
	var hdrs []byte
	if d.RequestHeaders != nil {
		hdrs, _ = json.Marshal(d.RequestHeaders)
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhook_deliveries (webhook_config_id, event_type, payload, request_url, request_headers, response_status, response_body, status, attempt_count, max_attempts, next_attempt_at, error_message, duration_ms, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())`,
		d.WebhookConfigID, d.EventType, d.Payload, d.RequestURL, hdrs,
		d.ResponseStatus, d.ResponseBody, d.Status, d.AttemptCount, d.MaxAttempts,
		d.NextAttemptAt, d.ErrorMessage, d.DurationMs)
	if err != nil {
		return fmt.Errorf("insert webhook_delivery: %w", err)
	}
	return nil
}

func (r *WebhookRepo) UpdateDelivery(ctx context.Context, d *model.WebhookDelivery) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE webhook_deliveries SET response_status=$1, response_body=$2, status=$3, attempt_count=$4, next_attempt_at=$5, error_message=$6, duration_ms=$7
		 WHERE id=$8`,
		d.ResponseStatus, d.ResponseBody, d.Status, d.AttemptCount, d.NextAttemptAt, d.ErrorMessage, d.DurationMs, d.ID)
	return err
}

func (r *WebhookRepo) GetDeliveryByID(ctx context.Context, id int64) (*model.WebhookDelivery, error) {
	var d model.WebhookDelivery
	var reqHdrs []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, webhook_config_id, event_type, payload, request_url, request_headers, response_status, response_body, status, attempt_count, max_attempts, next_attempt_at, error_message, duration_ms, created_at
		 FROM webhook_deliveries WHERE id = $1`, id,
	).Scan(&d.ID, &d.WebhookConfigID, &d.EventType, &d.Payload, &d.RequestURL, &reqHdrs,
		&d.ResponseStatus, &d.ResponseBody, &d.Status, &d.AttemptCount, &d.MaxAttempts,
		&d.NextAttemptAt, &d.ErrorMessage, &d.DurationMs, &d.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get webhook_delivery: %w", err)
	}
	if len(reqHdrs) > 0 {
		json.Unmarshal(reqHdrs, &d.RequestHeaders)
	}
	return &d, nil
}

func (r *WebhookRepo) ListDeliveriesByConfig(ctx context.Context, configID uuid.UUID, status string, limit int) ([]model.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT id, webhook_config_id, event_type, payload, request_url, request_headers, response_status, response_body, status, attempt_count, max_attempts, next_attempt_at, error_message, duration_ms, created_at
	      FROM webhook_deliveries WHERE webhook_config_id = $1`
	args := []any{configID}
	argN := 2
	if status != "" {
		q += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, status)
		argN++
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argN)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list webhook_deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []model.WebhookDelivery
	for rows.Next() {
		var d model.WebhookDelivery
		var reqHdrs []byte
		if err := rows.Scan(&d.ID, &d.WebhookConfigID, &d.EventType, &d.Payload, &d.RequestURL, &reqHdrs,
			&d.ResponseStatus, &d.ResponseBody, &d.Status, &d.AttemptCount, &d.MaxAttempts,
			&d.NextAttemptAt, &d.ErrorMessage, &d.DurationMs, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan webhook_delivery: %w", err)
		}
		if len(reqHdrs) > 0 {
			json.Unmarshal(reqHdrs, &d.RequestHeaders)
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

// --- helpers ---

func scanWebhookConfig(row pgx.Row) (*model.WebhookConfig, error) {
	var c model.WebhookConfig
	var hdrsRaw []byte
	err := row.Scan(&c.ID, &c.DeviceID, &c.Name, &c.URL, &c.Secret, &c.Events, &hdrsRaw, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan webhook_config: %w", err)
	}
	if len(hdrsRaw) > 0 {
		json.Unmarshal(hdrsRaw, &c.Headers)
	}
	return &c, nil
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
