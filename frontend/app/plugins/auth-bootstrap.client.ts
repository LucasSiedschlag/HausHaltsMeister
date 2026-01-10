export default defineNuxtPlugin({
  name: 'auth-bootstrap',
  setup() {
    const publicRoutes = ['/auth/login', '/auth/signup', '/auth/forgot-password', '/auth/oauth/callback']
    const route = useRoute()
    const { accessToken, user, refresh, me } = useAuth()
    const { preferences, fetchPreferences } = usePreferences()
    const { selectLedger } = useLedger()
    const router = useRouter()

    const resetUserScopedState = (options?: { keepPreferencesNeeded?: boolean }) => {
      useState<unknown[]>('ledgers_list', () => []).value = []
      useState<string | null>('active_ledger_id', () => null).value = null
      useState<string | null>('active_ledger_role', () => null).value = null
      useState<boolean>('ledger_context_loading', () => false).value = false
      useState<string | null>('ledger_context_error', () => null).value = null
      useState<Promise<unknown[]> | null>('ledger_context_pending', () => null).value = null

      useState<Record<string, unknown>>('ledger_members_by_ledger', () => ({})).value = {}
      useState<Record<string, boolean>>('ledger_members_loading', () => ({})).value = {}
      useState<Record<string, string | null>>('ledger_members_error', () => ({})).value = {}

      useState<unknown[]>('accounts_list', () => []).value = []
      useState<boolean>('accounts_loading', () => false).value = false
      useState<string | null>('accounts_error', () => null).value = null

      useState<unknown[]>('categories_list', () => []).value = []
      useState<boolean>('categories_loading', () => false).value = false
      useState<string | null>('categories_error', () => null).value = null

      useState<unknown[]>('journal_transactions', () => []).value = []
      useState<unknown | null>('journal_next_cursor', () => null).value = null
      useState<number>('journal_last_page_size', () => 0).value = 0
      useState<boolean>('journal_loading', () => false).value = false
      useState<string | null>('journal_error', () => null).value = null
      useState<boolean>('journal_create_open', () => false).value = false

      useState<unknown[]>('budget_versions', () => []).value = []
      useState<boolean>('budget_loading', () => false).value = false
      useState<string | null>('budget_error', () => null).value = null
      useState<unknown[]>('budget_editor_lines', () => []).value = []
      useState<boolean>('budget_editor_loading', () => false).value = false
      useState<boolean>('budget_editor_saving', () => false).value = false

      useState<unknown | null>('user_preferences', () => null).value = null
      useState<boolean>('preferences_loading', () => false).value = false
      useState<boolean>('preferences_saving', () => false).value = false
      const preferencesNeeded = useState<boolean>('preferences_needed', () => false)
      preferencesNeeded.value = options?.keepPreferencesNeeded ?? false

      const ledgerCookie = useCookie<string | null>('hhm_ledger_id', { sameSite: 'lax' })
      ledgerCookie.value = null
    }

    // Use onNuxtReady to run AFTER hydration completes
    onNuxtReady(async () => {
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

        if (!preferences.value) {
          await fetchPreferences()
        }

        const defaultLedgerId = preferences.value?.default_ledger_id
        if (defaultLedgerId) {
          try {
            await selectLedger(defaultLedgerId)
          } catch {
            // Fallback to manual selection if default ledger is unavailable.
          }
        }
      } catch {
        // On error, redirect to login
        await router.push('/auth/login')
      }
    })

    watch(
      () => user.value?.id ?? null,
      (nextId, prevId) => {
        if (prevId && nextId && nextId !== prevId) {
          resetUserScopedState({ keepPreferencesNeeded: true })
        } else if (prevId && !nextId) {
          resetUserScopedState()
        }
      },
      { immediate: true }
    )
  }
})
