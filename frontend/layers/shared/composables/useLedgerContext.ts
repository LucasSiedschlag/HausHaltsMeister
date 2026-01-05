import { computed } from 'vue'
import { isApiError } from '#layers/shared/utils/api-error'
import { useApiClient } from '#layers/shared/composables/useApiClient'
import { hasRoleRank } from '#layers/shared/utils/ledger-roles'
import type { LedgerRole } from '#layers/shared/utils/ledger-roles'
import { useAuth } from '#layers/auth/composables/useAuth'

export type LedgerSummary = {
  id: string
  name: string
  currency_code: string
  owner_user_id?: string
  created_at?: string
  updated_at?: string | null
  role?: LedgerRole
}

export type LedgerRoleInfo = {
  ledger_id: string
  role: LedgerRole
}

export const useLedgerContext = () => {
  const ledgers = useState<LedgerSummary[]>('ledgers_list', () => [])
  const activeLedgerId = useState<string | null>('active_ledger_id', () => null)
  const activeRole = useState<string | null>('active_ledger_role', () => null)
  const loading = useState<boolean>('ledger_context_loading', () => false)
  const error = useState<string | null>('ledger_context_error', () => null)
  const api = useApiClient()
  const pendingLoad = useState<Promise<LedgerSummary[]> | null>('ledger_context_pending', () => null)
  const { accessToken } = useAuth()

  const activeLedger = computed(() => {
    const id = activeLedgerId.value
    if (!id) return null
    return ledgers.value.find((ledger) => ledger.id === id) ?? null
  })

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const loadLedgers = async () => {
    // Never load ledgers during SSR - only on client
    if (import.meta.server) {
      return []
    }

    if (ledgers.value.length > 0) {
      return ledgers.value
    }
    if (pendingLoad.value) {
      return pendingLoad.value
    }
    if (!(await ensureAccessToken())) {
      return []
    }
    loading.value = true
    error.value = null
    const request = (async () => {
      try {
        const payload = await api<LedgerSummary[]>('/ledgers', {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${accessToken.value}`
          }
        })
        ledgers.value = payload
        return payload
      } catch (err) {
        handleLedgerError(err)
        return []
      } finally {
        loading.value = false
        pendingLoad.value = null
      }
    })()
    pendingLoad.value = request
    return request
  }

  const setActiveLedger = async (ledgerId: string) => {
    // Never load ledger data during SSR - only on client
    if (import.meta.server) {
      return
    }

    if (!(await ensureAccessToken())) {
      return
    }
    loading.value = true
    error.value = null
    try {
      const payload = await api<LedgerRoleInfo>(`/ledgers/${ledgerId}/me`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      activeLedgerId.value = ledgerId
      activeRole.value = payload.role

      // Save to cookie for persistence across page refreshes
      const ledgerCookie = useCookie<string | null>('hhm_ledger_id', { sameSite: 'lax' })
      ledgerCookie.value = ledgerId
    } catch (err) {
      handleLedgerError(err)
    } finally {
      loading.value = false
    }
  }

  const ensureLedger = async (ledgerId?: string) => {
    // Never ensure ledger during SSR - only on client
    if (import.meta.server) {
      return
    }

    const list = await loadLedgers()

    if (list.length === 0) {
      error.value = 'Nenhum ledger disponivel'
      return
    }

    if (ledgerId) {
      if (activeLedgerId.value === ledgerId && activeRole.value) {
        return
      }
      await setActiveLedger(ledgerId)
      return
    }

    // Only skip if we have both ledger AND role
    if (activeLedger.value && activeRole.value) {
      return
    }

    // Try to restore from cookie first
    const ledgerCookie = useCookie<string | null>('hhm_ledger_id', { sameSite: 'lax' })
    if (ledgerCookie.value && list.some((l) => l.id === ledgerCookie.value)) {
      await setActiveLedger(ledgerCookie.value)
      return
    }

    const candidate = list[0]
    if (!candidate) {
      error.value = 'Nenhum ledger disponivel'
      return
    }

    await setActiveLedger(candidate.id)
  }

  const handleLedgerError = (err: unknown) => {
    if (isApiError(err)) {
      error.value = err.message || 'Erro ao carregar ledger'
      return
    }
    error.value = 'Erro inesperado'
  }

  const hasRole = (role: LedgerRole) => hasRoleRank(activeRole.value as LedgerRole | null, role)

  return {
    ledgers,
    activeLedgerId,
    activeLedger,
    activeRole,
    loading,
    error,
    loadLedgers,
    setActiveLedger,
    ensureLedger,
    hasRole
  }
}
