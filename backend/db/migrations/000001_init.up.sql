CREATE TABLE users (
    id UUID PRIMARY KEY,
    phone_number TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password TEXT NOT NULL,
    password_error_count SMALLINT DEFAULT 0 NOT NULL,
    password_changed_at BIGINT,
    role_id SMALLINT NOT NULL,
    state TEXT NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE INDEX ON USERS (phone_number);

CREATE TABLE users_history (
    changed_at BIGINT NOT NULL,
    changed_type TEXT NOT NULL,
    changed_by UUID,
    changed_user_agent TEXT,
    changed_client_ip TEXT,
    user_id UUID NOT NULL,
    phone_number TEXT NOT NULL,
    name TEXT NOT NULL,
    password TEXT NOT NULL,
    password_error_count SMALLINT NOT NULL,
    password_changed_at BIGINT,
    role_id SMALLINT NOT NULL,
    state TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    history_created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE TABLE ver_codes (
    id UUID PRIMARY KEY,
    phone_number TEXT NOT NULL,
    code TEXT NOT NULL,
    type TEXT NOT NULL,
    is_blocked BOOLEAN DEFAULT false NOT NULL,
    request_id TEXT NOT NULL,
    expired_at BIGINT NOT NULL,
    create_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE TABLE tokens (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    client_ip TEXT NOT NULL,
    is_blocked BOOLEAN DEFAULT false NOT NULL,
    user_id UUID NOT NULL,
    expired_at BIGINT NOT NULL,
    issued_at BIGINT NOT NULL,
    create_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE TABLE stores (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    state TEXT NOT NULL,
    password TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE TABLE stores_history (
    changed_at BIGINT NOT NULL,
    changed_type TEXT NOT NULL,
    changed_by UUID,
    changed_user_agent TEXT,
    changed_client_ip TEXT,
    store_id UUID NOT NULL,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    state TEXT NOT NULL,
    password TEXT,
    created_at BIGINT NOT NULL,
    history_created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE TABLE store_users (
    store_id UUID NOT NULL,
    user_id UUID NOT NULL,
    balance INT NOT NULL DEFAULT 0,
    points INT NOT NULL DEFAULT 0,
    balance_earmark INT NOT NULL DEFAULT 0,
    points_earmark INT NOT NULL DEFAULT 0,
    role_id SMALLINT NOT NULL,
    state TEXT NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL,
    PRIMARY KEY (store_id, user_id)
);

CREATE TABLE store_users_history (
    changed_at BIGINT NOT NULL,
    changed_type TEXT NOT NULL,
    changed_by UUID,
    changed_user_agent TEXT,
    changed_client_ip TEXT,
    store_id UUID NOT NULL,
    user_id UUID NOT NULL,
    balance INT NOT NULL DEFAULT 0,
    points INT NOT NULL DEFAULT 0,
    balance_earmark INT NOT NULL DEFAULT 0,
    points_earmark INT NOT NULL DEFAULT 0,
    role_id SMALLINT NOT NULL,
    state TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    history_created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);


CREATE TABLE store_devices (
    store_id UUID NOT NULL,
    device_id TEXT NOT NULL,
    name TEXT NOT NULL,
    real_type TEXT NOT NULL,
    display_type TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL,
    PRIMARY KEY (store_id, device_id)
);

CREATE TABLE store_devices_history (
    changed_at BIGINT NOT NULL,
    changed_type TEXT NOT NULL,
    changed_by UUID,
    changed_user_agent TEXT,
    changed_client_ip TEXT,
    store_id UUID NOT NULL,
    device_id TEXT NOT NULL,
    name TEXT,
    real_type TEXT NOT NULL,
    display_type TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    history_created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE TABLE records (
    created_by UUID,
    created_user_agent TEXT,
    created_client_ip TEXT,
    type TEXT NOT NULL,
    store_id UUID NOT NULL,
    record_id TEXT,
    user_id UUID,
    device_id TEXT,
    from_online_payment TEXT,
    amount INT NOT NULL,
    point_amount INT,
    ts BIGINT NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);

CREATE UNIQUE INDEX ON records (store_id, record_id);
CREATE INDEX ON records (type);
CREATE INDEX ON records (device_id);
CREATE INDEX ON records (user_id);
CREATE INDEX ON records (created_by);