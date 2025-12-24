ALTER TABLE budget_items ADD CONSTRAINT budget_items_period_category_key UNIQUE (budget_period_id, category_id);
