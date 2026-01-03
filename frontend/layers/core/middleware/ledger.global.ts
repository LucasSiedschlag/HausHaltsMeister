import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'

export default defineNuxtRouteMiddleware(async (to) => {
  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']
  if (publicRoutes.includes(to.path) || to.path.startsWith('/ledgers')) {
    return
  }

  const { accessToken } = useAuth()
  if (!accessToken.value) {
    return
  }

  const { currentLedgerId, ledgers, fetchLedgers, selectLedger } = useLedger()

  if (currentLedgerId.value) return

  const fetched = await fetchLedgers()
  if (!currentLedgerId.value && fetched.length === 1) {
    const onlyLedger = fetched[0]
    if (onlyLedger) {
      selectLedger(onlyLedger.id)
      return
    }
  }

  if (!currentLedgerId.value) {
    return navigateTo('/ledgers')
  }
})
