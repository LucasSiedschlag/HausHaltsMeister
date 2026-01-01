export const useApiClient = () => {
  const headers = process.server ? useRequestHeaders(['cookie']) : undefined
  return $fetch.create({
    baseURL: '/api',
    headers,
    credentials: 'include'
  })
}
