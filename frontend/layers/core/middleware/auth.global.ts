export default defineNuxtRouteMiddleware(async (to) => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']
  const supportedLocales = ['pt-BR', 'en-US']
  const localeMatch = to.path.match(/^\/([^/]+)(?:\/|$)/)
  const currentLocale = localeMatch && supportedLocales.includes(localeMatch[1]) ? localeMatch[1] : null
  const normalizedPath = currentLocale ? to.path.replace(`/${currentLocale}`, '') || '/' : to.path

  if (publicRoutes.includes(normalizedPath)) {
    return
  }

  const { accessToken } = useAuth()
  if (!accessToken.value) {
    return navigateTo(currentLocale ? `/${currentLocale}/auth/login` : '/auth/login')
  }
})
