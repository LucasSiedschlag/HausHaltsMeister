import { computed } from 'vue'
import { isApiError } from '#layers/shared/utils/api-error'
import { useApiClient } from '#layers/shared/composables/useApiClient'
import { hasRoleRank } from '#layers/shared/utils/ledger-roles'
import type { LedgerRole } from '#layers/shared/utils/ledger-roles'

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

  const activeLedger = computed(() => {
    const id = activeLedgerId.value
    if (!id) return null
    return ledgers.value.find((ledger) => ledger.id === id) ?? null
  })

  const loadLedgers = async () => {
    if (ledgers.value.length > 0) {
      return ledgers.value
    }
    loading.value = true
    error.value = null
    try {
      const payload = await api<LedgerSummary[]>('/ledgers', { method: 'GET' })
      ledgers.value = payload
      return payload
    } catch (err) {
      handleLedgerError(err)
      return []
    } finally {
      loading.value = false
    }
  }

  const setActiveLedger = async (ledgerId: string) => {
    loading.value = true
    error.value = null
    try {
      const payload = await api<LedgerRoleInfo>(`/ledgers/${ledgerId}/me`, { method: 'GET' })
      activeLedgerId.value = ledgerId
      activeRole.value = payload.role
    } catch (err) {
      handleLedgerError(err)
    } finally {
      loading.value = false
    }
  }

  const ensureLedger = async (ledgerId?: string) => {
    const list = await loadLedgers()
    if (list.length === 0) {
      error.value = 'Nenhum ledger disponivel'
      return
    }

    if (ledgerId) {
      if (activeLedgerId.value === ledgerId) return
      await setActiveLedger(ledgerId)
      return
    }

    if (activeLedger.value) return

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
