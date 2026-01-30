-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS exchange_spreads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    residual_numerator BIGINT NOT NULL,
    residual_denominator BIGINT NOT NULL,
    target_currency VARCHAR(3) NOT NULL CHECK (target_currency IN ('USD', 'EUR')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_exchange_spreads_transaction_id ON exchange_spreads(transaction_id);
CREATE INDEX IF NOT EXISTS idx_exchange_spreads_created_at ON exchange_spreads(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS exchange_spreads CASCADE;
-- +goose StatementEnd

