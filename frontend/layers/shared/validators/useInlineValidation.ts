import { reactive, watch } from 'vue'
import type { Ref } from 'vue'
import type { FieldErrors, ValidationError } from './types'
import type { z } from 'zod'
import { validateField, validateForm } from './validate'

type RefRecord<T> = {
  [K in keyof T]: Ref<T[K]>
}

export const useInlineValidation = <T extends Record<string, unknown>>(
  fields: RefRecord<T>,
  schemas: { [K in keyof T]: z.ZodTypeAny }
) => {
  type FieldKey = keyof T & string
  const touched = reactive(
    Object.keys(fields).reduce((acc, key) => {
      acc[key as FieldKey] = false
      return acc
    }, {} as Record<FieldKey, boolean>)
  ) as Record<FieldKey, boolean>

  const errors = reactive(
    Object.keys(fields).reduce((acc, key) => {
      acc[key as FieldKey] = []
      return acc
    }, {} as FieldErrors<FieldKey>)
  ) as FieldErrors<FieldKey>

  const updateFieldError = (field: FieldKey) => {
    errors[field] = validateField(fields[field].value, schemas[field])
  }

  const touchField = (field: FieldKey) => {
    touched[field] = true
    updateFieldError(field)
  }

  const validateAll = () => {
    ;(Object.keys(fields) as FieldKey[]).forEach((key) => {
      touched[key] = true
      updateFieldError(key)
    })
    return Object.values(errors).some((fieldErrors) => fieldErrors.length > 0)
  }

  ;(Object.keys(fields) as FieldKey[]).forEach((key) => {
    watch(fields[key], () => {
      if (touched[key]) {
        updateFieldError(key)
      }
    })
  })

  const setErrors = (fieldErrors: FieldErrors<FieldKey>) => {
    ;(Object.keys(fieldErrors) as FieldKey[]).forEach((key) => {
      errors[key] = fieldErrors[key] as ValidationError[]
    })
  }

  const validateWithSchemas = () => {
    const values = Object.keys(fields).reduce((acc, key) => {
      acc[key as keyof T] = fields[key as keyof T].value
      return acc
    }, {} as T)
    const formErrors = validateForm(schemas, values)
    setErrors(formErrors)
    return Object.values(formErrors).some((fieldErrors) => fieldErrors.length > 0)
  }

  return {
    touched,
    errors,
    touchField,
    validateAll,
    validateWithSchemas
  }
}
