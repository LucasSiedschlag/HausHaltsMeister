export type PaymentMethodKind = 'CREDIT_CARD' | 'DEBIT_CARD' | 'CASH' | 'PIX' | 'BANK_SLIP'

export interface PaymentMethod {
  id: number
  name: string
  kind: PaymentMethodKind
  bank_name: string
  credit_limit?: number | null
  closing_day?: number | null
  due_day?: number | null
  is_active: boolean
}

export interface CreatePaymentMethodRequest {
  name: string
  kind: PaymentMethodKind
  bank_name: string
  credit_limit?: number | null
  closing_day?: number | null
  due_day?: number | null
}

export interface UpdatePaymentMethodRequest {
  name: string
  kind: PaymentMethodKind
  bank_name: string
  credit_limit?: number | null
  closing_day?: number | null
  due_day?: number | null
  is_active: boolean
}
