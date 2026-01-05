import { FetchError } from 'ofetch'
import type { ApiError, ApiErrorPayload } from '#layers/shared/types/api-error'

const unwrapPayload = (data: unknown): ApiErrorPayload => {
  if (!data || typeof data !== 'object') {
    return {}
  }
  const payload = data as ApiErrorPayload
  return payload
}

const pickCode = (payload: ApiErrorPayload): string =>
  payload.code ?? payload.error?.code ?? 'INTERNAL_SERVER_ERROR'

const pickMessage = (payload: ApiErrorPayload, fallback: string): string =>
  payload.message ?? payload.error?.message ?? fallback

export const normalizeApiError = (error: unknown, status?: number): ApiError => {
  if (isApiError(error)) {
    return error
  }

  const fetchError = error as FetchError | Error
  const payload = unwrapPayload(
    (fetchError as FetchError).data ?? (fetchError as Error).message
  )
  const fallbackMessage = fetchError instanceof Error ? fetchError.message : 'Erro inesperado'
  const normalized: ApiError = new Error(pickMessage(payload, fallbackMessage)) as ApiError
  normalized.code = pickCode(payload)
  const statusCode = status ?? (fetchError as FetchError).status ?? 500
  normalized.status = statusCode
  normalized.details = payload.details ?? payload.error?.details
  normalized.payload = payload
  return normalized
}

export const isApiError = (value: unknown): value is ApiError =>
  Boolean(value && typeof value === 'object' && 'status' in value && 'code' in value)
