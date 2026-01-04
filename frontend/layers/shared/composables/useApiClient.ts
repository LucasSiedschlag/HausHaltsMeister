import { normalizeApiError } from '#layers/shared/utils/api-error'

type ApiRequestOptions = Parameters<ReturnType<typeof $fetch.create>>[1] & {
  idempotencyKey?: string
}

export const useApiClient = () => {
  const headers = process.server ? useRequestHeaders(['cookie']) : undefined
  const client = $fetch.create({
    baseURL: '/api',
    headers,
    credentials: 'include'
  })

  return <T>(path: string, options?: ApiRequestOptions): Promise<T> => {
    if (!options) {
      return client<T>(path).catch((error) => {
        throw normalizeApiError(error)
      })
    }

    const { idempotencyKey, headers, ...rest } = options
    const mergedHeaders = new Headers(headers ?? {})
    if (idempotencyKey) {
      mergedHeaders.set('Idempotency-Key', idempotencyKey)
    }
    return client<T>(path, { ...rest, headers: mergedHeaders }).catch((error) => {
      throw normalizeApiError(error)
    })
  }
}
