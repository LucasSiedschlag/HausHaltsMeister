export type Direction = 'IN' | 'OUT'

export interface CashflowEntry {
  id: number
  date: string
  category_id: number
  category_name?: string
  payment_method_id: number
  payment_method_name?: string
  direction: Direction
  title: string
  amount: number
  is_fixed: boolean
  reversal_of_entry_id?: number | null
}

export interface CreateCashflowRequest {
  date: string
  category_id: number
  payment_method_id: number
  direction: Direction
  title: string
  amount: number
  is_fixed: boolean
}
