-- +goose Up
-- Seed default admin user: admin / admin123
-- Password is bcrypt hash of "admin123"
INSERT INTO app_users (id, username, password_hash, full_name, role, is_active, allowed_device_ids)
VALUES (
    gen_random_uuid(),
    'admin',
    '$2a$10$lGbQKQpntD44teSAMl2yMuGxs8EBlB5LGReq0hFFVJFukyfvCJHUm',
    'Administrator',
    'admin',
    true,
    '{}'
) ON CONFLICT (username) DO NOTHING;

-- +goose Down
DELETE FROM app_users WHERE username = 'admin';
