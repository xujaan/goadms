-- +goose Up
CREATE TABLE user_shifts (
    id BIGSERIAL PRIMARY KEY,
    fingerprint_user_id UUID REFERENCES fingerprint_users(id) ON DELETE CASCADE,
    shift_id UUID REFERENCES shifts(id) ON DELETE CASCADE,
    effective_date DATE NOT NULL,
    UNIQUE(fingerprint_user_id, shift_id, effective_date)
);

CREATE INDEX idx_user_shifts_fingerprint ON user_shifts(fingerprint_user_id);
CREATE INDEX idx_user_shifts_shift ON user_shifts(shift_id);

-- +goose Down
DROP TABLE IF EXISTS user_shifts;
