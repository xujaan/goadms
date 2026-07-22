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

type DeviceRepo struct {
	pool *pgxpool.Pool
}

func NewDeviceRepo(pool *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{pool: pool}
}

func (r *DeviceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Device, error) {
	d, err := r.scanDevice(ctx, `SELECT id, name, serial_number, ip_address, port, location, brand, protocol, timezone, handshake_config, last_handshake_at, is_active, created_at, updated_at FROM devices WHERE id = $1`, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return d, err
}

func (r *DeviceRepo) GetBySN(ctx context.Context, sn string) (*model.Device, error) {
	d, err := r.scanDevice(ctx, `SELECT id, name, serial_number, ip_address, port, location, brand, protocol, timezone, handshake_config, last_handshake_at, is_active, created_at, updated_at FROM devices WHERE serial_number = $1`, sn)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return d, err
}

func (r *DeviceRepo) List(ctx context.Context, activeOnly bool) ([]model.Device, error) {
	q := `SELECT id, name, serial_number, ip_address, port, location, brand, protocol, timezone, handshake_config, last_handshake_at, is_active, created_at, updated_at FROM devices`
	if activeOnly {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY name`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		d, err := scanDeviceRow(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, *d)
	}
	return devices, rows.Err()
}

func (r *DeviceRepo) Create(ctx context.Context, d *model.Device) error {
	d.ID = uuid.New()
	d.CreatedAt = time.Now()
	d.UpdatedAt = d.CreatedAt

	hc, _ := json.Marshal(d.HandshakeConfig)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO devices (id, name, serial_number, ip_address, port, location, brand, protocol, timezone, handshake_config, last_handshake_at, is_active, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		d.ID, d.Name, d.SerialNumber, d.IPAddress, d.Port, d.Location, d.Brand, d.Protocol, d.Timezone, hc, d.LastHandshakeAt, d.IsActive, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create device: %w", err)
	}
	return nil
}

func (r *DeviceRepo) Update(ctx context.Context, d *model.Device) error {
	d.UpdatedAt = time.Now()
	hc, _ := json.Marshal(d.HandshakeConfig)
	_, err := r.pool.Exec(ctx,
		`UPDATE devices SET name=$1, serial_number=$2, ip_address=$3, port=$4, location=$5, brand=$6, protocol=$7, timezone=$8, handshake_config=$9, last_handshake_at=$10, is_active=$11, updated_at=$12 WHERE id=$13`,
		d.Name, d.SerialNumber, d.IPAddress, d.Port, d.Location, d.Brand, d.Protocol, d.Timezone, hc, d.LastHandshakeAt, d.IsActive, d.UpdatedAt, d.ID)
	if err != nil {
		return fmt.Errorf("update device: %w", err)
	}
	return nil
}

func (r *DeviceRepo) UpsertBySN(ctx context.Context, d *model.Device) error {
	now := time.Now()
	hc, _ := json.Marshal(d.HandshakeConfig)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO devices (id, name, serial_number, ip_address, port, location, brand, protocol, timezone, handshake_config, last_handshake_at, is_active, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,true,$12,$12)
		 ON CONFLICT (serial_number) DO UPDATE SET
		   last_handshake_at = EXCLUDED.last_handshake_at,
		   ip_address = COALESCE(NULLIF(EXCLUDED.ip_address, ''), devices.ip_address),
		   updated_at = EXCLUDED.updated_at`,
		d.ID, d.Name, d.SerialNumber, d.IPAddress, d.Port, d.Location, d.Brand, d.Protocol, d.Timezone, hc, d.LastHandshakeAt, now)
	if err != nil {
		return fmt.Errorf("upsert device: %w", err)
	}
	return nil
}

func (r *DeviceRepo) RecordHandshake(ctx context.Context, sn string, ip string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO devices (id, name, serial_number, ip_address, port, protocol, timezone, handshake_config, last_handshake_at, is_active, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $1, $2, 4370, 'zk-tcp', 'Asia/Jakarta', '{}', $3, true, $3, $3)
		 ON CONFLICT (serial_number) DO UPDATE SET
		   last_handshake_at = EXCLUDED.last_handshake_at,
		   ip_address = COALESCE(NULLIF(EXCLUDED.ip_address, ''), devices.ip_address),
		   updated_at = EXCLUDED.updated_at`,
		sn, ip, now)
	if err != nil {
		return fmt.Errorf("record handshake: %w", err)
	}
	return nil
}

func (r *DeviceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete device: %w", err)
	}
	return nil
}

func (r *DeviceRepo) ListOffline(ctx context.Context, threshold time.Duration) ([]model.Device, error) {
	cutoff := time.Now().Add(-threshold)
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, serial_number, ip_address, port, location, brand, protocol, timezone, handshake_config, last_handshake_at, is_active, created_at, updated_at
		 FROM devices WHERE is_active = true AND (last_handshake_at IS NULL OR last_handshake_at < $1)`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("list offline: %w", err)
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		d, err := scanDeviceRow(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, *d)
	}
	return devices, rows.Err()
}

func (r *DeviceRepo) ListActiveForPull(ctx context.Context) ([]model.Device, error) {
	return r.List(ctx, true)
}

func (r *DeviceRepo) scanDevice(ctx context.Context, q string, args ...any) (*model.Device, error) {
	row := r.pool.QueryRow(ctx, q, args...)
	return scanDeviceRow(row)
}

func scanDeviceRow(row pgx.Row) (*model.Device, error) {
	var d model.Device
	var hc []byte
	err := row.Scan(&d.ID, &d.Name, &d.SerialNumber, &d.IPAddress, &d.Port, &d.Location, &d.Brand, &d.Protocol, &d.Timezone, &hc, &d.LastHandshakeAt, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan device: %w", err)
	}
	if len(hc) > 0 {
		json.Unmarshal(hc, &d.HandshakeConfig)
	}
	if d.HandshakeConfig.Stamp == 0 {
		d.HandshakeConfig = model.DefaultHandshakeConfig()
	}
	return &d, nil
}
