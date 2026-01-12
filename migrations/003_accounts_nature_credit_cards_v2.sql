-- Accounts: types + nature
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_type_check;

ALTER TABLE accounts
  ADD COLUMN IF NOT EXISTS nature varchar NOT NULL DEFAULT 'asset';

UPDATE accounts
SET type = 'wallet'
WHERE type = 'cash';

UPDATE accounts
SET type = 'current',
    nature = 'liability'
WHERE type = 'credit_card';

UPDATE accounts
SET nature = 'asset'
WHERE nature IS NULL;

ALTER TABLE accounts
  ADD CONSTRAINT accounts_type_check
    CHECK (type IN ('current', 'business', 'investment', 'exchange', 'wallet'));

ALTER TABLE accounts
  ADD CONSTRAINT accounts_nature_check
    CHECK (nature IN ('asset', 'liability'));

-- Credit cards: rebuild table with card_id + parent/liability accounts
CREATE TABLE credit_cards_new (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  parent_account_id uuid NOT NULL REFERENCES accounts(id),
  liability_account_id uuid NOT NULL REFERENCES accounts(id),
  label varchar,
  brand varchar NOT NULL DEFAULT 'other' REFERENCES card_networks(code),
  last4 char(4),
  cvv char(3),
  holder_name varchar,
  active boolean NOT NULL DEFAULT true,
  color varchar,
  style varchar,
  closing_day int NOT NULL,
  due_day int NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  CONSTRAINT credit_cards_style_check
    CHECK (style IS NULL OR style IN ('solid', 'gradient', 'glass', 'outline')),
  CONSTRAINT credit_cards_cvv_check
    CHECK (cvv IS NULL OR cvv ~ '^[0-9]{3}$'),
  CONSTRAINT credit_cards_closing_day_check
    CHECK (closing_day BETWEEN 1 AND 31),
  CONSTRAINT credit_cards_due_day_check
    CHECK (due_day BETWEEN 1 AND 31)
);

INSERT INTO credit_cards_new (
  id,
  ledger_id,
  parent_account_id,
  liability_account_id,
  label,
  brand,
  last4,
  closing_day,
  due_day,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  a.ledger_id,
  COALESCE(pa.id, a.id),
  c.account_id,
  COALESCE(c.nickname, c.issuer_name),
  c.network,
  c.last4,
  c.closing_day,
  c.due_day,
  c.created_at,
  c.updated_at
FROM credit_cards c
JOIN accounts a ON a.id = c.account_id
LEFT JOIN LATERAL (
  SELECT pa.id
  FROM accounts pa
  WHERE pa.ledger_id = a.ledger_id
    AND pa.nature = 'asset'
  ORDER BY pa.created_at
  LIMIT 1
) pa ON true;

DROP TABLE credit_cards;
ALTER TABLE credit_cards_new RENAME TO credit_cards;

CREATE INDEX credit_cards_ledger_id_idx ON credit_cards (ledger_id);
CREATE INDEX credit_cards_parent_account_idx ON credit_cards (ledger_id, parent_account_id);
CREATE INDEX credit_cards_liability_account_idx ON credit_cards (ledger_id, liability_account_id);

-- Installment plans: move to credit_card_id
ALTER TABLE installment_plans
  ADD COLUMN credit_card_id uuid;

UPDATE installment_plans ip
SET credit_card_id = cc.id
FROM credit_cards cc
WHERE cc.liability_account_id = ip.card_account_id
  AND cc.ledger_id = ip.ledger_id;

ALTER TABLE installment_plans
  ALTER COLUMN credit_card_id SET NOT NULL;

ALTER TABLE installment_plans
  ADD CONSTRAINT installment_plans_credit_card_id_fkey
    FOREIGN KEY (credit_card_id) REFERENCES credit_cards(id);

CREATE INDEX installment_plans_credit_card_id_idx ON installment_plans (credit_card_id);

DROP INDEX IF EXISTS installment_plans_card_account_id_idx;
ALTER TABLE installment_plans DROP COLUMN card_account_id;

-- Statements: move to credit_card_id
ALTER TABLE credit_card_statements
  ADD COLUMN credit_card_id uuid;

UPDATE credit_card_statements cs
SET credit_card_id = cc.id
FROM credit_cards cc
WHERE cc.liability_account_id = cs.card_account_id
  AND cc.ledger_id = cs.ledger_id;

ALTER TABLE credit_card_statements
  ALTER COLUMN credit_card_id SET NOT NULL;

ALTER TABLE credit_card_statements
  ADD CONSTRAINT credit_card_statements_credit_card_id_fkey
    FOREIGN KEY (credit_card_id) REFERENCES credit_cards(id);

ALTER TABLE credit_card_statements
  DROP CONSTRAINT IF EXISTS credit_card_statements_card_account_id_statement_month_key;

ALTER TABLE credit_card_statements
  DROP COLUMN card_account_id;

ALTER TABLE credit_card_statements
  ADD CONSTRAINT credit_card_statements_credit_card_id_statement_month_key
    UNIQUE (credit_card_id, statement_month);

-- Transactions: link to credit_card_id (optional for legacy rows)
ALTER TABLE transactions
  ADD COLUMN credit_card_id uuid;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_credit_card_id_fkey
    FOREIGN KEY (credit_card_id) REFERENCES credit_cards(id);

CREATE INDEX transactions_credit_card_id_idx ON transactions (ledger_id, credit_card_id);
