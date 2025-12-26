export type BudgetMode = 'ABSOLUTE' | 'PERCENT_OF_INCOME'

export interface BudgetItem {
  id: number
  budget_period_id: number
  category_id: number
  category_name?: string
  mode: BudgetMode
  planned_amount: number
  actual_amount: number
  target_percent: number
}

export interface BudgetSummary {
  month: string
  total_income: number
  items: BudgetItem[]
}

export interface SetBudgetItemRequest {
  category_id: number
  mode: BudgetMode
  planned_amount?: number
  target_percent?: number
}

export interface UpdateBudgetItemRequest {
  mode: BudgetMode
  planned_amount?: number
  target_percent?: number
}

export interface CategoryOption {
  id: number
  name: string
  direction: 'IN' | 'OUT'
  is_active: boolean
  is_budget_relevant: boolean
  inactive_from_month?: string
}
