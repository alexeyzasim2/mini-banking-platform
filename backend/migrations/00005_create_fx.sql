-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts
  ADD COLUMN IF NOT EXISTS allow_negative BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE accounts
  DROP CONSTRAINT IF EXISTS accounts_balance_cents_check;

ALTER TABLE accounts
  ADD CONSTRAINT accounts_balance_cents_check
  CHECK (balance_cents >= 0 OR allow_negative = TRUE);


INSERT INTO users (id, email, password, first_name, last_name)
VALUES ('00000000-0000-0000-0000-000000000001', 'fx@system.local', 'N/A', 'FX', 'System')
ON CONFLICT (email) DO NOTHING;


INSERT INTO accounts (user_id, currency, balance_cents, allow_negative)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'USD', 0, TRUE),
  ('00000000-0000-0000-0000-000000000001', 'EUR', 0, TRUE)
ON CONFLICT (user_id, currency) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts DROP COLUMN IF EXISTS allow_negative;
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_balance_cents_check;
-- +goose StatementEnd