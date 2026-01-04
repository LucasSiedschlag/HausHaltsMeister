CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- NOTE: All enumerations are enforced via CHECK constraints below.

-- Core
CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email varchar NOT NULL UNIQUE,
  display_name varchar,
  avatar_url varchar,
  email_verified_at timestamptz,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz
);

CREATE TABLE user_preferences (
  user_id uuid PRIMARY KEY REFERENCES users(id),
  theme_mode varchar NOT NULL DEFAULT 'system',
  theme_palette varchar NOT NULL DEFAULT 'default',
  theme_tone varchar NOT NULL DEFAULT 'vivid',
  locale varchar NOT NULL DEFAULT 'pt-BR',
  compact_mode varchar NOT NULL DEFAULT 'comfortable',
  font_scale varchar NOT NULL DEFAULT 'md',
  notify_card_close boolean NOT NULL DEFAULT true,
  notify_budget_over boolean NOT NULL DEFAULT true,
  notify_payables boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: theme and locale values.
  CONSTRAINT user_preferences_theme_mode_check CHECK (theme_mode IN ('light', 'dark', 'system')),
  CONSTRAINT user_preferences_theme_palette_check CHECK (theme_palette IN ('default', 'red', 'rose', 'orange', 'green', 'yellow', 'violet', 'monochrome')),
  CONSTRAINT user_preferences_theme_tone_check CHECK (theme_tone IN ('vivid', 'pastel', 'muted')),
  CONSTRAINT user_preferences_locale_check CHECK (locale IN ('pt-BR', 'en-US')),
  CONSTRAINT user_preferences_font_scale_check CHECK (font_scale IN ('sm', 'md', 'lg')),
  CONSTRAINT user_preferences_compact_mode_check CHECK (compact_mode IN ('comfortable', 'compact', 'dense'))
);

-- Auth (identities, secrets, sessions, oauth states)
CREATE TABLE auth_secrets (
  user_id uuid PRIMARY KEY REFERENCES users(id),
  password_hash varchar NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz
);

CREATE TABLE auth_identities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id),
  provider varchar NOT NULL,
  provider_user_id varchar NOT NULL,
  email varchar,
  display_name varchar,
  avatar_url varchar,
  email_verified boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  last_login_at timestamptz,
  -- Validation: provider allowed values (extend when adding providers).
  CONSTRAINT auth_identities_provider_check
    CHECK (provider IN ('password', 'google', 'github')),
  UNIQUE (provider, provider_user_id),
  UNIQUE (user_id, provider)
);

CREATE TABLE auth_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id),
  refresh_token_hash varchar NOT NULL,
  is_persistent boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  user_agent varchar,
  ip varchar,
  device_name varchar,
  rotated_from_session_id uuid REFERENCES auth_sessions(id),
  UNIQUE (refresh_token_hash)
);

CREATE TABLE oauth_states (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  provider varchar NOT NULL,
  state varchar NOT NULL,
  code_verifier varchar NOT NULL,
  redirect_uri varchar,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  expires_at timestamptz NOT NULL,
  used_at timestamptz,
  -- Validation: provider allowed values (extend when adding providers).
  CONSTRAINT oauth_states_provider_check
    CHECK (provider IN ('google', 'github')),
  UNIQUE (state)
);

CREATE TABLE ledgers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id uuid NOT NULL REFERENCES users(id),
  name varchar NOT NULL,
  currency_code varchar NOT NULL DEFAULT 'BRL',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz
);

CREATE TABLE ledger_members (
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  user_id uuid NOT NULL REFERENCES users(id),
  role varchar NOT NULL DEFAULT 'viewer',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  removed_at timestamptz,
  -- Validation: ledger role allowed values.
  CONSTRAINT ledger_members_role_check CHECK (role IN ('owner', 'editor', 'viewer')),
  UNIQUE (ledger_id, user_id)
);

CREATE TABLE audit_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  user_id uuid NOT NULL REFERENCES users(id),
  action varchar NOT NULL,
  entity_id uuid,
  ip varchar,
  user_agent varchar,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz
);

CREATE INDEX audit_log_ledger_created_at_idx ON audit_log (ledger_id, created_at);

CREATE TABLE accounts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  name varchar NOT NULL,
  type varchar NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: account type allowed values.
  CONSTRAINT accounts_type_check CHECK (type IN ('cash', 'investment', 'credit_card')),
  UNIQUE (ledger_id, name)
);

CREATE TABLE categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  parent_id uuid REFERENCES categories(id),
  name varchar NOT NULL,
  direction varchar NOT NULL,
  is_budget_base boolean NOT NULL DEFAULT false,
  is_budget_relevant boolean NOT NULL DEFAULT true,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: category direction allowed values.
  CONSTRAINT categories_direction_check CHECK (direction IN ('in', 'out')),
  -- Validation: IN categories cannot be budget relevant.
  CONSTRAINT categories_in_budget_relevant_check
    CHECK (direction <> 'in' OR is_budget_relevant = false),
  -- Validation: OUT categories cannot be budget base.
  CONSTRAINT categories_out_budget_base_check
    CHECK (direction <> 'out' OR is_budget_base = false),
  UNIQUE (ledger_id, name)
);

CREATE TABLE transactions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  occurred_at timestamptz NOT NULL,
  description varchar NOT NULL,
  notes text,
  created_by_user_id uuid NOT NULL REFERENCES users(id),
  external_source varchar,
  external_id varchar,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  UNIQUE (ledger_id, external_source, external_id)
);

CREATE TABLE idempotency_keys (
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  key varchar(200) NOT NULL,
  resource_type varchar NOT NULL,
  resource_id uuid NOT NULL,
  request_hash varchar(64) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (ledger_id, key)
);

CREATE TABLE entries (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  transaction_id uuid NOT NULL REFERENCES transactions(id),
  account_id uuid NOT NULL REFERENCES accounts(id),
  category_id uuid REFERENCES categories(id),
  kind varchar NOT NULL DEFAULT 'normal',
  amount_cents bigint NOT NULL,
  memo text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: entry kind allowed values.
  CONSTRAINT entries_kind_check CHECK (kind IN ('normal', 'transfer', 'adjust')),
  -- Validation: positive amounts only.
  CONSTRAINT entries_amount_cents_check CHECK (amount_cents > 0)
);

-- Budget
CREATE TABLE budget_plans (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  name varchar NOT NULL DEFAULT 'Default',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  UNIQUE (ledger_id)
);

CREATE TABLE budget_plan_versions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  plan_id uuid NOT NULL REFERENCES budget_plans(id),
  effective_from_month date NOT NULL,
  created_by_user_id uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: always first day of month.
  CONSTRAINT budget_plan_versions_month_check
    CHECK (date_trunc('month', effective_from_month)::date = effective_from_month),
  UNIQUE (plan_id, effective_from_month)
);

CREATE TABLE budget_plan_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  version_id uuid NOT NULL REFERENCES budget_plan_versions(id),
  category_id uuid NOT NULL REFERENCES categories(id),
  percent numeric NOT NULL,
  include_children boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  UNIQUE (version_id, category_id)
);

-- Credit cards
CREATE TABLE card_networks (
  code varchar PRIMARY KEY,
  display_name varchar NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz
);

INSERT INTO card_networks (code, display_name) VALUES
  ('visa', 'Visa'),
  ('mastercard', 'Mastercard'),
  ('elo', 'Elo'),
  ('amex', 'American Express'),
  ('hipercard', 'Hipercard'),
  ('other', 'Other');

CREATE TABLE credit_cards (
  account_id uuid PRIMARY KEY REFERENCES accounts(id),
  issuer_name varchar,
  network varchar NOT NULL DEFAULT 'other' REFERENCES card_networks(code),
  nickname varchar,
  last4 char(4),
  credit_limit_cents bigint,
  closing_day int NOT NULL,
  due_day int NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: credit limit cannot be negative.
  CONSTRAINT credit_cards_credit_limit_check
    CHECK (credit_limit_cents IS NULL OR credit_limit_cents >= 0),
  -- Validation: closing day must be 1..31.
  CONSTRAINT credit_cards_closing_day_check
    CHECK (closing_day BETWEEN 1 AND 31),
  -- Validation: due day must be 1..31.
  CONSTRAINT credit_cards_due_day_check
    CHECK (due_day BETWEEN 1 AND 31)
);

CREATE TABLE installment_plans (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  card_account_id uuid NOT NULL REFERENCES accounts(id),
  purchase_occurred_at timestamptz NOT NULL,
  merchant varchar,
  description varchar NOT NULL,
  category_id uuid NOT NULL REFERENCES categories(id),
  total_amount_cents bigint NOT NULL,
  installments_count int NOT NULL,
  installment_amount_cents bigint NOT NULL,
  first_due_month date NOT NULL,
  status varchar NOT NULL DEFAULT 'active',
  created_by_user_id uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: installment plan status allowed values.
  CONSTRAINT installment_plans_status_check
    CHECK (status IN ('active', 'cancelled', 'finished')),
  -- Validation: always first day of month.
  CONSTRAINT installment_plans_month_check
    CHECK (date_trunc('month', first_due_month)::date = first_due_month)
);

CREATE TABLE credit_card_statements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  card_account_id uuid NOT NULL REFERENCES accounts(id),
  statement_month date NOT NULL,
  closing_date date NOT NULL,
  due_date date NOT NULL,
  total_charges_cents bigint NOT NULL DEFAULT 0,
  total_payments_cents bigint NOT NULL DEFAULT 0,
  status varchar NOT NULL DEFAULT 'open',
  payment_transaction_id uuid REFERENCES transactions(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: statement status allowed values.
  CONSTRAINT credit_card_statements_status_check
    CHECK (status IN ('open', 'closed', 'paid')),
  -- Validation: always first day of month.
  CONSTRAINT credit_card_statements_month_check
    CHECK (date_trunc('month', statement_month)::date = statement_month),
  UNIQUE (card_account_id, statement_month)
);

CREATE TABLE installments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ledger_id uuid NOT NULL REFERENCES ledgers(id),
  plan_id uuid NOT NULL REFERENCES installment_plans(id),
  installment_no int NOT NULL,
  due_month date NOT NULL,
  amount_cents bigint NOT NULL,
  status varchar NOT NULL DEFAULT 'scheduled',
  posted_transaction_id uuid REFERENCES transactions(id),
  paid_statement_id uuid REFERENCES credit_card_statements(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz,
  -- Validation: installment status allowed values.
  CONSTRAINT installments_status_check
    CHECK (status IN ('scheduled', 'posted', 'paid', 'skipped')),
  -- Validation: amount must be positive.
  CONSTRAINT installments_amount_check CHECK (amount_cents > 0),
  -- Validation: always first day of month.
  CONSTRAINT installments_month_check
    CHECK (date_trunc('month', due_month)::date = due_month),
  UNIQUE (plan_id, installment_no)
);

-- Indexes
CREATE INDEX auth_identities_user_id_idx ON auth_identities (user_id);
CREATE INDEX auth_identities_provider_idx ON auth_identities (provider);

CREATE INDEX auth_sessions_user_id_idx ON auth_sessions (user_id);
CREATE INDEX auth_sessions_expires_at_idx ON auth_sessions (expires_at);
CREATE INDEX auth_sessions_revoked_at_idx ON auth_sessions (revoked_at);

CREATE INDEX oauth_states_provider_idx ON oauth_states (provider);
CREATE INDEX oauth_states_expires_at_idx ON oauth_states (expires_at);

CREATE INDEX accounts_ledger_id_idx ON accounts (ledger_id);

CREATE INDEX categories_ledger_id_idx ON categories (ledger_id);
CREATE INDEX categories_parent_id_idx ON categories (parent_id);

CREATE INDEX transactions_ledger_id_occurred_at_idx ON transactions (ledger_id, occurred_at);

CREATE INDEX entries_ledger_transaction_idx ON entries (ledger_id, transaction_id);
CREATE INDEX entries_ledger_account_idx ON entries (ledger_id, account_id);
CREATE INDEX entries_ledger_category_idx ON entries (ledger_id, category_id);

CREATE INDEX installment_plans_ledger_id_idx ON installment_plans (ledger_id);
CREATE INDEX installment_plans_card_account_id_idx ON installment_plans (card_account_id);
CREATE INDEX installment_plans_ledger_first_due_month_idx ON installment_plans (ledger_id, first_due_month);

CREATE INDEX credit_card_statements_ledger_id_idx ON credit_card_statements (ledger_id);

CREATE INDEX installments_ledger_due_month_idx ON installments (ledger_id, due_month);
CREATE INDEX installments_status_idx ON installments (status);
