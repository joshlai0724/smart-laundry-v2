CREATE TABLE records (
    id UUID PRIMARY KEY,
    device_id TEXT NOT NULL,
    type TEXT NOT NULL,
    amount INT NOT NULL,
    is_uploaded BOOLEAN NOT NULL,
    uploaded_at BIGINT,
    ts BIGINT NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 NOT NULL
);