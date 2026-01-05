import { normalizeApiError } from '#layers/shared/utils/api-error'

type ApiRequestOptions = Parameters<ReturnType<typeof $fetch.create>>[1] & {
  idempotencyKey?: string
}

export const useApiClient = () => {
  const config = useRuntimeConfig()
  const headers = import.meta.server ? useRequestHeaders(['cookie']) : undefined
  const baseHeaders = headers ? new Headers(headers) : undefined
  const client = $fetch.create({
    baseURL: import.meta.server ? config.apiBase : config.public.apiBase,
    headers,
    credentials: 'include',
    onResponse({ response }) {
      if (!import.meta.server) return
      const event = useRequestEvent()
      if (!event) return
      const responseHeaders = response.headers as Headers | Record<string, string | string[] | undefined>
      let cookies: string[] | string | null | undefined
      if ('getSetCookie' in responseHeaders && typeof responseHeaders.getSetCookie === 'function') {
        cookies = responseHeaders.getSetCookie()
      } else if ('get' in responseHeaders && typeof responseHeaders.get === 'function') {
        cookies = responseHeaders.get('set-cookie')
      } else if (!('get' in responseHeaders)) {
        cookies = responseHeaders['set-cookie']
      }
      if (cookies) {
        appendResponseHeader(event, 'set-cookie', cookies)
      }
    }
  })

  return <T>(path: string, options?: ApiRequestOptions): Promise<T> => {
    if (!options) {
      return client<T>(path).catch((error) => {
        throw normalizeApiError(error)
      })
    }

    const { idempotencyKey, headers, ...rest } = options
    const mergedHeaders = baseHeaders ? new Headers(baseHeaders) : new Headers()
    if (headers) {
      const provided = new Headers(headers as HeadersInit)
      provided.forEach((value, key) => {
        mergedHeaders.set(key, value)
      })
    }
    if (idempotencyKey) {
      mergedHeaders.set('Idempotency-Key', idempotencyKey)
    }
    return client<T>(path, { ...rest, headers: mergedHeaders }).catch((error) => {
      throw normalizeApiError(error)
    })
  }
}
