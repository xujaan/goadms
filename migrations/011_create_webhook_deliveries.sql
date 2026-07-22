-- +goose Up
CREATE TABLE webhook_deliveries (
    id BIGSERIAL PRIMARY KEY,
    webhook_config_id UUID REFERENCES webhook_configs(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    request_url TEXT NOT NULL,
    request_headers JSONB,
    response_status INTEGER,
    response_body TEXT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending','success','failed','retrying')),
    attempt_count INTEGER DEFAULT 1,
    max_attempts INTEGER DEFAULT 5,
    next_attempt_at TIMESTAMPTZ,
    error_message TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_webhook_deliveries_config ON webhook_deliveries(webhook_config_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status, next_attempt_at);

-- +goose Down
DROP TABLE IF EXISTS webhook_deliveries;
