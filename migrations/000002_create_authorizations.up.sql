CREATE TYPE authorization_status AS ENUM ('APPROVED', 'DECLINED', 'REVERSED');

CREATE TABLE authorizations (
    id                  UUID                    PRIMARY KEY DEFAULT uuid_generate_v4(),
    authorization_code  VARCHAR(20)             NOT NULL UNIQUE,
    card_id             UUID                    NOT NULL REFERENCES cards(id),
    merchant_id         VARCHAR(255)            NOT NULL,
    merchant_name       VARCHAR(255)            NOT NULL,
    amount              NUMERIC(19, 4)          NOT NULL,
    currency            VARCHAR(3)              NOT NULL,
    status              authorization_status    NOT NULL,
    created_at          TIMESTAMP               NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_authorizations_card_id ON authorizations(card_id);
CREATE INDEX idx_authorizations_authorization_code ON authorizations(authorization_code);