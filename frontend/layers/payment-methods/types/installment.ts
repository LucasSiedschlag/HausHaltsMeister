export type InstallmentAmountMode = 'TOTAL' | 'INSTALLMENT'

export interface CreateInstallmentRequest {
  description: string
  amount_mode?: InstallmentAmountMode
  total_amount?: number
  installment_amount?: number
  count: number
  category_id: number
  payment_method_id: number
  purchase_date: string
}

export interface InstallmentPlanResponse {
  id: number
  description: string
  total_amount: number
  installment_count: number
  installment_amount: number
  start_month: string
  payment_method_id: number
}

export interface InvoiceEntry {
  date: string
  title: string
  amount: number
  category_name: string
}

export interface InvoiceSummary {
  month: string
  total: number
  total_remaining: number
  entries: InvoiceEntry[]
}

export interface InstallmentCategoryOption {
  id: number
  name: string
  direction: string
  is_budget_relevant: boolean
  is_active: boolean
}
