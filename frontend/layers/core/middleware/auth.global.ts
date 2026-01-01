export default defineNuxtRouteMiddleware((to) => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password']
  if (publicRoutes.includes(to.path)) {
    return
  }

  const accessToken = useState<string | null>('access_token', () => null)
  if (!accessToken.value) {
    return navigateTo('/auth/login')
  }
})
