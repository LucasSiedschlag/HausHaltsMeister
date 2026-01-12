import { z } from 'zod'
import { messages } from './messages'
import type { ValidationError } from './types'

type AddIssue = (code: string, message: string) => void

const issue = (ctx: z.RefinementCtx, code: string, message: string) => {
  ctx.addIssue({
    code: z.ZodIssueCode.custom,
    message,
    params: { code }
  })
}

const handleRequired = (value: string, add: AddIssue, label: string) => {
  if (!value.trim()) {
    add('required', messages.required(label))
    return true
  }
  return false
}

export const emailSchema = (label = 'E-mail') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (value.includes(' ')) {
      add('email_no_spaces', messages.emailNoSpaces())
      return
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      add('email_invalid', messages.emailInvalid())
      return
    }
    if (value.length > 254) {
      add('max_length', messages.maxLength(label, 254))
    }
  })

export const passwordSchema = (label = 'Senha') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (value.length < 8) {
      add('min_length', messages.minLength(label, 8))
      return
    }
    if (!/[A-Z]/.test(value)) {
      add('password_uppercase', messages.passwordUppercase())
      return
    }
    if (!/[0-9]/.test(value)) {
      add('password_number', messages.passwordNumber())
    }
  })

export const nameSchema = (label = 'Nome') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (value.trim().length < 2) {
      add('min_length', messages.minLength(label, 2))
      return
    }
    if (value.trim().length > 80) {
      add('max_length', messages.maxLength(label, 80))
    }
  })

export const descriptionSchema = (label = 'Descricao') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (value.trim().length < 2) {
      add('min_length', messages.minLength(label, 2))
      return
    }
    if (value.trim().length > 120) {
      add('max_length', messages.maxLength(label, 120))
    }
  })

export const dateSchema = (label = 'Data') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (!/^[0-9]{4}-[0-9]{2}-[0-9]{2}$/.test(value)) {
      add('invalid_date', messages.invalidDate(label))
    }
  })

export const amountSchema = (label = 'Valor') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    const digits = value.replace(/\D/g, '')
    if (!digits || Number(digits) <= 0) {
      add('min_value', messages.amountPositive(label))
    }
  })

export const integerSchema = (
  label = 'Numero',
  options: { min?: number; max?: number } = {}
) =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    const trimmed = value.trim()
    if (!/^\d+$/.test(trimmed)) {
      add('invalid_number', messages.invalidNumber(label))
      return
    }
    const parsed = Number(trimmed)
    if (options.min !== undefined && parsed < options.min) {
      add('min_value', messages.minValue(label, options.min))
    }
    if (options.max !== undefined && parsed > options.max) {
      add('max_value', messages.maxValue(label, options.max))
    }
  })

export const monthSchema = (label = 'Mes') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (!/^[0-9]{4}-[0-9]{2}$/.test(value)) {
      add('invalid_date', messages.invalidDate(label))
    }
  })

export const codeSchema = (label = 'Codigo') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (value.trim().length < 2) {
      add('min_length', messages.minLength(label, 2))
      return
    }
    if (value.trim().length > 20) {
      add('max_length', messages.maxLength(label, 20))
      return
    }
    if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value)) {
      add('invalid_code', messages.invalidCode(label))
    }
  })

export const accountTypeSchema = (label = 'Tipo') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (!['current', 'business', 'investment', 'exchange', 'wallet'].includes(value)) {
      add('invalid', messages.invalidOption(label))
    }
  })

export const categoryDirectionSchema = (label = 'Direção') =>
  z.string().superRefine((value, ctx) => {
    const add: AddIssue = (code, message) => issue(ctx, code, message)
    if (handleRequired(value, add, label)) return
    if (!['in', 'out'].includes(value)) {
      add('invalid', messages.invalidOption(label))
    }
  })

export const issuesFromSchema = (schema: z.ZodTypeAny, value: unknown): ValidationError[] => {
  const result = schema.safeParse(value)
  if (result.success) return []
  return result.error.issues.map((issueItem) => {
    const params = (issueItem as z.ZodIssue & { params?: { code?: string } }).params
    return {
      code: params?.code || issueItem.code,
      message: issueItem.message
    }
  })
}
