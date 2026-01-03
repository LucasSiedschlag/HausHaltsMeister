import { normalizeApiError } from '#layers/shared/utils/api-error'

export const useApiClient = () => {
  const headers = process.server ? useRequestHeaders(['cookie']) : undefined
  const client = $fetch.create({
    baseURL: '/api',
    headers,
    credentials: 'include'
  })

  return <T>(path: string, options?: Parameters<typeof client>[1]): Promise<T> =>
    client<T>(path, options).catch((error) => {
      throw normalizeApiError(error)
    })
}
