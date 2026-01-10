import { useLedger } from '@shared/composables/useLedger'
import { usePreferences } from '@shared/composables/usePreferences'
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

  const { currentLedgerId, fetchLedgers, selectLedger } = useLedger()
  const { preferences, fetchPreferences } = usePreferences()

  if (currentLedgerId.value) return

  const fetched = await fetchLedgers()
  if (!currentLedgerId.value && fetched.length === 1) {
    const onlyLedger = fetched[0]
    if (onlyLedger) {
      await selectLedger(onlyLedger.id)
      return
    }
  }

  if (!currentLedgerId.value) {
    if (!preferences.value) {
      try {
        await fetchPreferences()
      } catch {
        // Ignore preference load errors here.
      }
    }
    const defaultLedgerId = preferences.value?.default_ledger_id
    if (defaultLedgerId) {
      try {
        await selectLedger(defaultLedgerId)
        return
      } catch {
        // Ignore invalid default ledger and continue to selector.
      }
    }
  }

  if (!currentLedgerId.value) {
    return navigateTo('/ledgers')
  }
})
