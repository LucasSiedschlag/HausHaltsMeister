import type { FieldErrors, ValidationError } from './types'
import type { z } from 'zod'
import { issuesFromSchema } from './schemas'

type ValidateOptions = {
  firstOnly?: boolean
}

export const validateField = (
  value: unknown,
  schema: z.ZodTypeAny,
  options: ValidateOptions = { firstOnly: true }
): ValidationError[] => {
  const errors = issuesFromSchema(schema, value)
  if (options.firstOnly) {
    return errors.slice(0, 1)
  }
  return errors
}

export const validateForm = <T extends Record<string, unknown>>(
  schemas: { [K in keyof T]: z.ZodTypeAny },
  values: T,
  options: ValidateOptions = { firstOnly: true }
): FieldErrors<keyof T & string> => {
  const fieldErrors = {} as FieldErrors<keyof T & string>
  ;(Object.keys(schemas) as Array<keyof T>).forEach((key) => {
    fieldErrors[key as keyof typeof fieldErrors] = validateField(values[key], schemas[key], options)
  })
  return fieldErrors
}
