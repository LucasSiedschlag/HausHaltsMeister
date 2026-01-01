INSERT INTO categories (
  ledger_id,
  name,
  direction,
  is_budget_base,
  is_budget_relevant,
  is_active
)
SELECT
  l.id,
  seed.name,
  seed.direction,
  seed.is_budget_base,
  seed.is_budget_relevant,
  true
FROM ledgers AS l
CROSS JOIN (
  VALUES
    ('Pagamento Fatura Cartão', 'out', false, false),
    ('Entrada Pagamento Cartão', 'in', false, false),
    ('Aportes Investimentos', 'out', false, false),
    ('Aporte Investimentos', 'in', false, false),
    ('Resgate Investimentos', 'out', false, false),
    ('Resgate Investimentos', 'in', false, false),
    ('Rendimentos', 'in', false, false)
) AS seed(name, direction, is_budget_base, is_budget_relevant)
ON CONFLICT (ledger_id, name) DO NOTHING;
