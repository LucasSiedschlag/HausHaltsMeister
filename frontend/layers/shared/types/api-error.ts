export interface ApiErrorPayload {
  code?: string
  message?: string
  details?: Record<string, string>
  error?: ApiErrorPayload
}

export interface ApiError extends Error {
  status: number
  code: string
  details?: Record<string, string>
  payload?: ApiErrorPayload
}
