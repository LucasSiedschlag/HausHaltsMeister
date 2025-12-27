ALTER TABLE installment_plans
  ADD COLUMN interest_rate_unit varchar(20),
  ADD COLUMN starts_on_current_invoice boolean NOT NULL DEFAULT true;

ALTER TABLE installment_plans
  DROP CONSTRAINT IF EXISTS installment_plans_plan_type_check;

ALTER TABLE installment_plans
  ADD CONSTRAINT installment_plans_plan_type_check
  CHECK (plan_type IN ('ONE_OFF', 'INSTALLMENT', 'RECURRING', 'CARD_INSTALLMENT'));

ALTER TABLE installment_plans
  DROP CONSTRAINT IF EXISTS installment_plans_payment_method_id_check;

ALTER TABLE installment_plan_items
  ADD COLUMN extra_amount decimal(14,2) NOT NULL DEFAULT 0,
  ADD COLUMN is_paid boolean NOT NULL DEFAULT false,
  ADD COLUMN paid_at timestamptz;
