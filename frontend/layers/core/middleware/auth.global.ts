export default defineNuxtRouteMiddleware(async (to) => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']
  if (publicRoutes.includes(to.path)) {
    return
  }

  const { accessToken } = useAuth()
  if (!accessToken.value) {
    return navigateTo('/auth/login')
  }
})
