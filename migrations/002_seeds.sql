INSERT INTO flow_categories (name, direction, is_budget_relevant, is_active) VALUES
-- Entradas (IN)
('Salário', 'IN', true, true),
('Investimento', 'IN', true, true), -- Rendimento/Resgate
('Ganho Extra', 'IN', true, true),
('Estorno', 'IN', false, true),

-- Saídas (OUT) - Orçamento
('Investimentos', 'OUT', true, true), -- Aporte destinado a investimentos
('Metas', 'OUT', true, true),
('Conhecimento', 'OUT', true, true),
('Custos fixos', 'OUT', true, true),
('Conforto', 'OUT', true, true),
('Prazeres', 'OUT', true, true),

-- Saídas (OUT) - Fora do Orçamento (Picuinhas/Técnicas)
('Picuinhas', 'OUT', false, true),
('Fatura Cartão', 'OUT', false, true);
