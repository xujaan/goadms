-- +goose Up
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    port INTEGER DEFAULT 4370,
    location VARCHAR(255),
    brand VARCHAR(100),
    protocol VARCHAR(20) DEFAULT 'zk-tcp' CHECK (protocol IN ('zk-tcp', 'adms-http')),
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    handshake_config JSONB DEFAULT '{}',
    last_handshake_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS devices;
