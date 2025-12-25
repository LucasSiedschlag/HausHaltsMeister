import type { CreateCategoryRequest, Direction } from '../types/category'

export interface CategoryFormErrors {
  name?: string
  direction?: string
}

const allowedDirections: Direction[] = ['IN', 'OUT']

export function validateCategoryInput(input: CreateCategoryRequest) {
  const errors: CategoryFormErrors = {}

  if (!input.name.trim()) {
    errors.name = 'Informe um nome para a categoria.'
  }

  if (!allowedDirections.includes(input.direction)) {
    errors.direction = 'Selecione uma direcao valida.'
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
  }
}
