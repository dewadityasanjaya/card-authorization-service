CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE card_status AS ENUM ('ACTIVE', 'FROZEN');

CREATE TABLE cards (
    id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    card_number     VARCHAR(16)     NOT NULL UNIQUE,
    cardholder_name VARCHAR(255)    NOT NULL,
    status          card_status     NOT NULL DEFAULT 'ACTIVE',
    currency        VARCHAR(3)      NOT NULL,
    balance         NUMERIC(19, 4)  NOT NULL DEFAULT 0,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cards_card_number ON cards(card_number);