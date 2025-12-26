import type { CreatePaymentMethodRequest, PaymentMethodKind } from '../types/payment-method'

export interface PaymentMethodFormInput {
  name: string
  kind: string
  bank_name: string
  credit_limit: string
  closing_day: string
  due_day: string
}

export interface PaymentMethodFormErrors {
  name?: string
  kind?: string
  credit_limit?: string
  closing_day?: string
  due_day?: string
}

export interface PaymentMethodFormValues extends CreatePaymentMethodRequest {}

const validKinds: PaymentMethodKind[] = ['CREDIT_CARD', 'DEBIT_CARD', 'CASH', 'PIX', 'BANK_SLIP']

function parseDay(value: string, field: 'closing_day' | 'due_day', errors: PaymentMethodFormErrors) {
  const trimmed = value.trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  if (!Number.isInteger(parsed) || parsed < 1 || parsed > 31) {
    errors[field] = 'Informe um dia válido entre 1 e 31.'
    return undefined
  }
  return parsed
}

export function validatePaymentMethodInput(input: PaymentMethodFormInput) {
  const errors: PaymentMethodFormErrors = {}

  const name = input.name.trim()
  if (!name) {
    errors.name = 'Informe o nome do meio de pagamento.'
  }

  const kind = input.kind as PaymentMethodKind
  if (!input.kind) {
    errors.kind = 'Selecione um tipo.'
  } else if (!validKinds.includes(kind)) {
    errors.kind = 'Selecione um tipo válido.'
  }

  const isCreditCard = kind === 'CREDIT_CARD'
  let creditLimit: number | undefined
  if (input.credit_limit.trim()) {
    const normalizedLimit = input.credit_limit.trim().replace(',', '.')
    const parsedLimit = Number(normalizedLimit)
    if (Number.isNaN(parsedLimit) || parsedLimit <= 0) {
      errors.credit_limit = 'Informe um limite válido.'
    } else {
      creditLimit = parsedLimit
    }
  }
  const closingDay = isCreditCard ? parseDay(input.closing_day, 'closing_day', errors) : undefined
  const dueDay = isCreditCard ? parseDay(input.due_day, 'due_day', errors) : undefined

  return {
    valid: Object.keys(errors).length === 0,
    errors,
    values: {
      name,
      kind: kind || 'CREDIT_CARD',
      bank_name: input.bank_name.trim(),
      credit_limit: creditLimit,
      closing_day: closingDay,
      due_day: dueDay,
    } as PaymentMethodFormValues,
  }
}
