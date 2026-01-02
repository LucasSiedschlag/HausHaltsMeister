export type ValidationError = {
  code: string
  message: string
}

export type FieldErrors<T extends string = string> = Record<T, ValidationError[]>

export type Rule<T> = (value: T) => ValidationError | null
