export type Direction = 'IN' | 'OUT'

export interface Category {
  id: number
  name: string
  direction: Direction
  is_budget_relevant: boolean
  is_active: boolean
}

export interface CreateCategoryRequest {
  name: string
  direction: Direction
  is_budget_relevant: boolean
}
