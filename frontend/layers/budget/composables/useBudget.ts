import { useApiClient } from '@shared/composables/useApiClient'
import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'

export type BudgetVersion = {
  id: string
  ledger_id: string
  effective_from_month: string // YYYY-MM-01
  is_active: boolean
  created_at?: string
}

export type BudgetVersionCreatePayload = {
  effective_from_month: string
  lines?: { category_id: string; percent: number; include_children: boolean }[]
}

export const useBudget = () => {
  const api = useApiClient()
  const { currentLedgerId } = useLedger()
  const { accessToken } = useAuth()

  const versions = useState<BudgetVersion[]>('budget_versions', () => [])
  const loading = useState<boolean>('budget_loading', () => false)
  const error = useState<string | null>('budget_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchVersions = async () => {
    if (!(await ensureAccessToken())) return []
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return []

    loading.value = true
    error.value = null
    try {
      const payload = await api<BudgetVersion[]>(`/ledgers/${ledgerId}/budget/versions`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${accessToken.value}` }
      })
      versions.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar versões do orçamento'
      return []
    } finally {
      loading.value = false
    }
  }

  const createVersion = async (payload: BudgetVersionCreatePayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null

    try {
      const created = await api<BudgetVersion>(`/ledgers/${ledgerId}/budget/versions`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${accessToken.value}` },
        body: payload
      })
      versions.value = [created, ...versions.value]
      return created
    } catch (err) {
      throw err
    }
  }

  const deleteVersion = async (versionId: string) => {
    if (!(await ensureAccessToken())) return false
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return false

    try {
      await api(`/ledgers/${ledgerId}/budget/versions/${versionId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${accessToken.value}` }
      })
      versions.value = versions.value.filter((v) => v.id !== versionId)
      return true
    } catch (err) {
      throw err
    }
  }

  return {
    versions,
    loading,
    error,
    fetchVersions,
    createVersion,
    deleteVersion
  }
}
