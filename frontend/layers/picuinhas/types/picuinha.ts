export type PicuinhaKind = 'PLUS' | 'MINUS'
export type CardOwner = 'SELF' | 'THIRD'
export type PicuinhaCaseType = 'ONE_OFF' | 'INSTALLMENT' | 'RECURRING' | 'CARD_INSTALLMENT'
export type PicuinhaCaseStatus = 'OPEN' | 'PAID' | 'RECURRING'

export interface Person {
  id: number
  name: string
  notes: string
  balance: number
}

export interface CreatePersonRequest {
  name: string
  notes: string
}

export interface UpdatePersonRequest {
  name: string
  notes: string
}

export interface PicuinhaEntry {
  id: number
  person_id: number
  amount: number
  kind: PicuinhaKind
  cash_flow_id?: number | null
  payment_method_id?: number | null
  card_owner: CardOwner
  created_at: string
}

export interface AddEntryRequest {
  person_id: number
  amount: number
  kind: PicuinhaKind
  auto_create_flow: boolean
  payment_method_id?: number | null
  card_owner?: CardOwner
}

export interface PaymentMethod {
  id: number
  name: string
  kind: string
  bank_name: string
  closing_day?: number | null
  due_day?: number | null
  is_active: boolean
}

export interface PicuinhaCase {
  id: number
  person_id: number
  title: string
  case_type: PicuinhaCaseType
  total_amount?: number | null
  installment_count?: number | null
  installment_amount?: number | null
  start_date: string
  payment_method_id?: number | null
  installment_plan_id?: number | null
  category_id?: number | null
  interest_rate?: number | null
  interest_rate_unit?: string | null
  recurrence_interval_months?: number | null
  installments_total: number
  installments_paid: number
  amount_paid: number
  amount_remaining: number
  status: PicuinhaCaseStatus
}

export interface PicuinhaCaseInstallment {
  id: number
  case_id: number
  installment_number: number
  due_date: string
  amount: number
  extra_amount: number
  is_paid: boolean
  paid_at?: string | null
}

export interface CreatePicuinhaCaseRequest {
  person_id: number
  title: string
  case_type: PicuinhaCaseType
  total_amount?: number
  installment_count?: number
  installment_amount?: number
  start_date: string
  payment_method_id?: number | null
  installment_plan_id?: number | null
  category_id?: number | null
  interest_rate?: number | null
  interest_rate_unit?: string | null
  recurrence_interval_months?: number | null
}

export interface UpdatePicuinhaInstallmentRequest {
  is_paid: boolean
  extra_amount: number
}
