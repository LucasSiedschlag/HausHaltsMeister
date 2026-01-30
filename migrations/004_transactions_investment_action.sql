ALTER TABLE transactions
  ADD COLUMN investment_action varchar;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_investment_action_check
    CHECK (investment_action IS NULL OR investment_action IN ('contribution', 'redemption', 'earnings', 'loss'));

CREATE INDEX transactions_investment_action_idx
  ON transactions (ledger_id, investment_action);
