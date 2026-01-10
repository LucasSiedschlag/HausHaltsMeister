export default defineNuxtPlugin(() => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']
  const { sessionExpired } = useAuth()
  const { blocked } = useSessionBlock()
  const route = useRoute()

  watch(
    () => [sessionExpired.value, route.path],
    ([expired, path]) => {
      blocked.value = Boolean(expired && !publicRoutes.includes(path))
    },
    { immediate: true }
  )
})
