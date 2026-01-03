import { computed } from 'vue'
import type { ApiError } from '#layers/shared/types/api-error'
import { isApiError } from '#layers/shared/utils/api-error'
import { useApiClient } from '#layers/shared/composables/useApiClient'

export type LedgerSummary = {
  id: string
  name: string
  currency_code: string
  owner_user_id?: string
}

export type LedgerRoleInfo = {
  ledger_id: string
  role: 'viewer' | 'editor' | 'owner'
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

  const hasRole = (role: 'viewer' | 'editor' | 'owner') =>
    activeRole.value ? roleRank(activeRole.value) >= roleRank(role) : false

  const roleRank = (role: string) => {
    switch (role) {
      case 'owner':
        return 3
      case 'editor':
        return 2
      case 'viewer':
        return 1
      default:
        return 0
    }
  }

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
