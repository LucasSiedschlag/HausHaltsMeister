export interface BudgetFormInput {
  category_id: string
  planned_amount: string
}

export interface BudgetFormErrors {
  category_id?: string
  planned_amount?: string
}

export interface BudgetFormValues {
  category_id: number
  planned_amount: number
}

export function validateBudgetItemInput(input: BudgetFormInput) {
  const errors: BudgetFormErrors = {}

  const categoryId = Number(input.category_id)
  if (!input.category_id || Number.isNaN(categoryId) || categoryId <= 0) {
    errors.category_id = 'Selecione uma categoria.'
  }

  const normalizedAmount = input.planned_amount.trim().replace(',', '.')
  const amount = Number(normalizedAmount)
  if (!normalizedAmount) {
    errors.planned_amount = 'Informe um valor planejado.'
  } else if (Number.isNaN(amount)) {
    errors.planned_amount = 'Informe um valor válido.'
  } else if (amount < 0) {
    errors.planned_amount = 'O valor não pode ser negativo.'
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
    values: {
      category_id: categoryId,
      planned_amount: amount,
    } as BudgetFormValues,
  }
}
