export type Direction = 'IN' | 'OUT'

export interface Category {
  id: number
  name: string
  direction: Direction
  is_budget_relevant: boolean
  is_active: boolean
  inactive_from_month?: string
}

export interface CreateCategoryRequest {
  name: string
  direction: Direction
  is_budget_relevant: boolean
}

export interface UpdateCategoryRequest {
  name: string
  direction: Direction
  is_budget_relevant: boolean
  is_active: boolean
}
