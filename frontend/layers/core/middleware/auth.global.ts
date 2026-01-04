export default defineNuxtRouteMiddleware(async (to) => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']

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

  const { accessToken, refresh } = useAuth()
  if (!accessToken.value) {
    try {
      await refresh()
    } catch {
      return navigateTo('/auth/login')
    }
  }
  if (!accessToken.value) {
    return navigateTo('/auth/login')
  }
})
