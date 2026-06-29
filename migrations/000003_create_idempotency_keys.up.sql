CREATE TABLE idempotency_keys (
    key                 VARCHAR(255)    PRIMARY KEY,
    authorization_id    UUID            NOT NULL REFERENCES authorizations(id),
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);