export interface ErrorResponse {
  error: string
  request_id?: string
}

export interface ApiError extends Error {
  statusCode?: number
  requestId?: string
  data?: ErrorResponse
}

export function getApiErrorMessage(error: unknown) {
  const err = error as ApiError
  if (err?.data?.error) {
    return err.data.error
  }
  if (err?.message) {
    return err.message
  }
  return 'Unexpected error'
}

export function useApiClient() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBaseUrl || '/api'
  const token = config.public.apiToken

  const request = async <T>(path: string, options: any = {}): Promise<T> => {
    const headers = new Headers(options.headers || {})
    if (token) {
      headers.set('X-App-Token', token)
    }

    try {
      return await $fetch<T>(path, {
        baseURL,
        ...options,
        headers,
      })
    } catch (error: any) {
      const apiError: ApiError = new Error(getApiErrorMessage(error))
      apiError.statusCode = error?.statusCode
      apiError.data = error?.data
      apiError.requestId = error?.data?.request_id
      throw apiError
    }
  }

  return { request }
}
