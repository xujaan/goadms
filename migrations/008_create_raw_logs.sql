-- +goose Up
CREATE TABLE raw_logs (
    id BIGSERIAL PRIMARY KEY,
    device_sn VARCHAR(100),
    request_method VARCHAR(10),
    request_uri TEXT,
    query_params JSONB,
    request_body TEXT,
    response_body TEXT,
    log_type VARCHAR(30) CHECK (log_type IN ('push_handshake','push_records','push_test','push_getrequest','pull_attendance','device_action')),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_raw_logs_device_sn ON raw_logs(device_sn);
CREATE INDEX idx_raw_logs_type ON raw_logs(log_type);
CREATE INDEX idx_raw_logs_created ON raw_logs(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS raw_logs;
