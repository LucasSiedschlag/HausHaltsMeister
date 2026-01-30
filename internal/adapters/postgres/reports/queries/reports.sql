-- Reports

-- name: ListAccountBalances :many
SELECT a.id, a.name, a.type, a.nature,
  COALESCE(SUM(
    CASE
      WHEN t.id IS NULL THEN 0
      WHEN t.investment_action IS NULL THEN
        CASE
          WHEN c.direction = 'in' THEN e.amount_cents
          WHEN c.direction = 'out' THEN -e.amount_cents
          ELSE 0
        END
      WHEN t.investment_action = 'contribution' THEN
        CASE WHEN a.type = 'investment' THEN e.amount_cents ELSE -e.amount_cents END
      WHEN t.investment_action = 'redemption' THEN
        CASE WHEN a.type = 'investment' THEN -e.amount_cents ELSE e.amount_cents END
      WHEN t.investment_action = 'earnings' THEN
        CASE WHEN a.type = 'investment' THEN e.amount_cents ELSE 0 END
      WHEN t.investment_action = 'loss' THEN
        CASE WHEN a.type = 'investment' THEN -e.amount_cents ELSE 0 END
      ELSE 0
    END
  ), 0) AS net_cents
FROM accounts a
LEFT JOIN entries e ON e.account_id = a.id AND e.ledger_id = a.ledger_id
LEFT JOIN transactions t ON t.id = e.transaction_id AND t.occurred_at < $2
LEFT JOIN categories c ON c.id = e.category_id
WHERE a.ledger_id = $1
GROUP BY a.id, a.name, a.type, a.nature
ORDER BY a.created_at;

-- name: ListCategorySummary :many
SELECT c.id, c.name, c.direction,
  COALESCE(SUM(
    CASE
      WHEN t.id IS NULL THEN 0
      ELSE e.amount_cents
    END
  ), 0) AS total_cents
FROM categories c
LEFT JOIN entries e ON e.category_id = c.id AND e.ledger_id = c.ledger_id
LEFT JOIN transactions t ON t.id = e.transaction_id AND t.occurred_at >= $2 AND t.occurred_at <= $3
WHERE c.ledger_id = $1
GROUP BY c.id, c.name, c.direction
ORDER BY c.name;

-- name: ListCashflow :many
SELECT date_trunc('month', t.occurred_at)::date AS month,
  COALESCE(SUM(
    CASE
      WHEN t.investment_action IS NULL THEN
        CASE WHEN c.direction = 'in' THEN e.amount_cents ELSE 0 END
      WHEN t.investment_action = 'redemption' AND a.type <> 'investment' THEN e.amount_cents
      ELSE 0
    END
  ), 0) AS total_in,
  COALESCE(SUM(
    CASE
      WHEN t.investment_action IS NULL THEN
        CASE WHEN c.direction = 'out' THEN e.amount_cents ELSE 0 END
      WHEN t.investment_action = 'contribution' AND a.type <> 'investment' THEN e.amount_cents
      ELSE 0
    END
  ), 0) AS total_out
FROM entries e
JOIN transactions t ON t.id = e.transaction_id
JOIN accounts a ON a.id = e.account_id
LEFT JOIN categories c ON c.id = e.category_id
WHERE e.ledger_id = $1 AND t.occurred_at >= $2 AND t.occurred_at <= $3
GROUP BY month
ORDER BY month;
