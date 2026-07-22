-- +goose Up
CREATE TABLE attendances (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    device_sn VARCHAR(100) NOT NULL,
    employee_id VARCHAR(100) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    status1 SMALLINT DEFAULT 0,
    status2 SMALLINT DEFAULT 0,
    status3 SMALLINT DEFAULT 0,
    status4 SMALLINT DEFAULT 0,
    status5 SMALLINT DEFAULT 0,
    source VARCHAR(10) NOT NULL CHECK (source IN ('push', 'pull')),
    raw_payload TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX idx_attendances_dedup ON attendances(
    device_sn, employee_id, timestamp,
    COALESCE(status1, -1), COALESCE(status2, -1),
    COALESCE(status3, -1), COALESCE(status4, -1), COALESCE(status5, -1)
);
CREATE INDEX idx_attendances_device_sn ON attendances(device_sn);
CREATE INDEX idx_attendances_employee_id ON attendances(employee_id);
CREATE INDEX idx_attendances_timestamp ON attendances(timestamp DESC);
CREATE INDEX idx_attendances_device_time ON attendances(device_id, timestamp DESC);

-- +goose Down
DROP TABLE IF EXISTS attendances;
