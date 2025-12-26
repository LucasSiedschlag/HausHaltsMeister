import type { CreateInstallmentRequest, InstallmentAmountMode } from '../types/installment'

export interface InstallmentFormInput {
  description: string
  amount_mode: InstallmentAmountMode | ''
  total_amount: string
  installment_amount: string
  count: string
  category_id: string
  payment_method_id: string
  purchase_date: string
}

export interface InstallmentFormErrors {
  description?: string
  amount_mode?: string
  total_amount?: string
  installment_amount?: string
  count?: string
  category_id?: string
  payment_method_id?: string
  purchase_date?: string
}

export interface InstallmentFormValues extends CreateInstallmentRequest {}

function parseAmount(value: string) {
  const normalized = value.trim().replace(',', '.')
  if (!normalized) return { value: 0, valid: false }
  const amount = Number(normalized)
  if (Number.isNaN(amount)) return { value: 0, valid: false }
  return { value: amount, valid: true }
}

export function validateInstallmentInput(input: InstallmentFormInput) {
  const errors: InstallmentFormErrors = {}

  if (!input.description.trim()) {
    errors.description = 'Informe a descrição da compra.'
  }

  const mode = input.amount_mode as InstallmentAmountMode
  if (!mode) {
    errors.amount_mode = 'Selecione como deseja informar o valor.'
  }

  const countValue = Number(input.count)
  if (!input.count.trim()) {
    errors.count = 'Informe a quantidade de parcelas.'
  } else if (Number.isNaN(countValue) || countValue < 1 || !Number.isInteger(countValue)) {
    errors.count = 'Informe um número inteiro de parcelas.'
  }

  const categoryId = Number(input.category_id)
  if (!input.category_id || Number.isNaN(categoryId) || categoryId <= 0) {
    errors.category_id = 'Selecione uma categoria.'
  }

  const paymentMethodId = Number(input.payment_method_id)
  if (!input.payment_method_id || Number.isNaN(paymentMethodId) || paymentMethodId <= 0) {
    errors.payment_method_id = 'Selecione um cartão.'
  }

  if (!input.purchase_date) {
    errors.purchase_date = 'Informe a data da compra.'
  } else if (Number.isNaN(new Date(input.purchase_date).getTime())) {
    errors.purchase_date = 'Informe uma data válida.'
  }

  let totalAmount: number | undefined
  let installmentAmount: number | undefined

  if (mode === 'TOTAL') {
    const parsed = parseAmount(input.total_amount)
    if (!parsed.valid || parsed.value <= 0) {
      errors.total_amount = 'Informe o valor total.'
    } else {
      totalAmount = parsed.value
    }
  }

  if (mode === 'INSTALLMENT') {
    const parsed = parseAmount(input.installment_amount)
    if (!parsed.valid || parsed.value <= 0) {
      errors.installment_amount = 'Informe o valor da parcela.'
    } else {
      installmentAmount = parsed.value
    }
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
    values: {
      description: input.description.trim(),
      amount_mode: mode,
      total_amount: totalAmount,
      installment_amount: installmentAmount,
      count: Number.isNaN(countValue) ? 0 : countValue,
      category_id: categoryId,
      payment_method_id: paymentMethodId,
      purchase_date: input.purchase_date,
    } as InstallmentFormValues,
  }
}
