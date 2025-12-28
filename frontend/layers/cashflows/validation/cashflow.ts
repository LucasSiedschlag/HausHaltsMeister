export interface CashflowFormErrors {
  title?: string
  amount?: string
  category_id?: string
  payment_method_id?: string
  date?: string
}

export interface CashflowFormValues {
  title: string
  amount: number
  category_id: number
  payment_method_id: number
  date?: string
}

interface ValidationInput {
  title: string
  amount: string
  category_id: string
  payment_method_id: string
  date?: string
}

export function validateCashflowInput(input: ValidationInput, options: { requiresDate: boolean }) {
  const errors: CashflowFormErrors = {}

  if (!input.title.trim()) {
    errors.title = 'Informe um título.'
  }

  const amountValue = Number(input.amount)
  if (!amountValue || amountValue <= 0) {
    errors.amount = 'Informe um valor válido.'
  }

  const categoryId = Number(input.category_id)
  if (!categoryId) {
    errors.category_id = 'Selecione uma categoria.'
  }

  const paymentMethodId = Number(input.payment_method_id)
  if (!paymentMethodId) {
    errors.payment_method_id = 'Selecione um meio de pagamento.'
  }

  if (options.requiresDate && !input.date) {
    errors.date = 'Informe a data da compra.'
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
    values: {
      title: input.title.trim(),
      amount: amountValue,
      category_id: categoryId,
      payment_method_id: paymentMethodId,
      date: input.date,
    } as CashflowFormValues,
  }
}
