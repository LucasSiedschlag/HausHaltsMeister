ALTER TABLE user_preferences
ADD COLUMN default_ledger_id uuid REFERENCES ledgers(id);
