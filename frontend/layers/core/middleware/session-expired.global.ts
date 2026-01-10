import { abortNavigation } from '#app'

export default defineNuxtRouteMiddleware((to) => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']
  const { sessionExpired } = useAuth()

  if (publicRoutes.includes(to.path)) {
    return
  }

  if (sessionExpired.value) {
    return abortNavigation()
  }
})
