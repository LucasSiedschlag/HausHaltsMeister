export default defineNuxtRouteMiddleware(async (to) => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']

  // Handle locale detection on public routes
  if (publicRoutes.includes(to.path)) {
    const { accessToken } = useAuth()
    if (!accessToken.value) {
      const localeCookie = useCookie<string | null>('hhm_locale')
      if (!localeCookie.value) {
        const normalizeLocale = (value: string) => {
          const lower = value.toLowerCase()
          if (lower.startsWith('en')) return 'en-US'
          return 'pt-BR'
        }
        const detectLocale = () => {
          if (import.meta.server) {
            const headers = useRequestHeaders(['accept-language'])
            const header = headers['accept-language'] || ''
            const primary = header.split(',')[0]?.trim()
            return normalizeLocale(primary || 'pt-BR')
          }
          if (typeof navigator !== 'undefined') {
            const primary = navigator.languages?.[0] || navigator.language || 'pt-BR'
            return normalizeLocale(primary)
          }
          return 'pt-BR'
        }
        const detected = detectLocale()
        localeCookie.value = detected
        const { $i18n } = useNuxtApp()
        if ($i18n?.setLocale) {
          await $i18n.setLocale(detected)
        }
      }
    }
    return
  }

  // Protected routes: check auth
  // On SSR: just check if token exists (will be refreshed client-side if needed)
  // On client: refresh will be handled by the auth-bootstrap plugin
  const { accessToken } = useAuth()

  if (import.meta.server) {
    // On SSR, just check if we have a refresh cookie
    // Don't try to refresh here - let the client handle it
    const cookies = useRequestHeaders(['cookie'])
    const hasRefreshCookie = cookies.cookie?.includes('hhm_refresh')
    if (!hasRefreshCookie) {
      return navigateTo('/auth/login')
    }
    return
  }

  // On client: plugin will handle refresh before navigation
  // Middleware just needs to exist for SSR check above
})
