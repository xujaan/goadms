-- +goose Up
CREATE TABLE device_users (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    fingerprint_user_id UUID REFERENCES fingerprint_users(id) ON DELETE CASCADE,
    uid INTEGER NOT NULL,
    employee_code VARCHAR(100),
    full_name VARCHAR(255),
    privilege INTEGER DEFAULT 0,
    fingerprint_count INTEGER DEFAULT 0,
    synced_at TIMESTAMPTZ,
    UNIQUE(device_id, uid),
    UNIQUE(device_id, employee_code)
);

CREATE INDEX idx_device_users_device ON device_users(device_id);
CREATE INDEX idx_device_users_fingerprint ON device_users(fingerprint_user_id);

-- +goose Down
DROP TABLE IF EXISTS device_users;
