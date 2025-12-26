export type PicuinhaKind = 'PLUS' | 'MINUS'
export type CardOwner = 'SELF' | 'THIRD'

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
