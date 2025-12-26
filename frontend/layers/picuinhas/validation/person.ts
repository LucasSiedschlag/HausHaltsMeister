import type { CreatePersonRequest } from '../types/picuinha'

export interface PersonFormErrors {
  name?: string
}

export function validatePersonInput(input: CreatePersonRequest) {
  const errors: PersonFormErrors = {}

  if (!input.name.trim()) {
    errors.name = 'Informe o nome da pessoa.'
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
  }
}
