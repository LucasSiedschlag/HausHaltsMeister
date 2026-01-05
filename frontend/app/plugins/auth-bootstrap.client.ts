export default defineNuxtPlugin({
  name: 'auth-bootstrap',
  setup() {
    const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']

    // Use onNuxtReady to run AFTER hydration completes
    onNuxtReady(async () => {
      const route = useRoute()
      const { accessToken, user, refresh, me } = useAuth()
      const router = useRouter()

      // Public routes don't need auth
      if (publicRoutes.includes(route.path)) {
        return
      }

      try {
        // Refresh if no token
        if (!accessToken.value) {
          await refresh()
        }

        // Ensure we have user data
        if (!user.value) {
          await me()
        }

        // Note: Ledger selection is handled by user preferences (default_ledger_id)
        // or manually via LedgerSwitcher component
      } catch {
        // On error, redirect to login
        await router.push('/auth/login')
      }
    })
  }
})
