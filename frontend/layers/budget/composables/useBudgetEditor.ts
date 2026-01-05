import { useApiClient } from '@shared/composables/useApiClient'
import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'
import { debounce } from 'perfect-debounce'

export type BudgetLine = {
  id: string
  version_id: string
  category_id: string
  percent: number // Integer 0-100
  include_children: boolean
  created_at?: string
  updated_at?: string
}

export type BudgetLineUpdatePayload = {
  percent?: number
  include_children?: boolean
}

export const useBudgetEditor = () => {
  const api = useApiClient()
  const { currentLedgerId } = useLedger()
  const { accessToken } = useAuth()

  const lines = useState<BudgetLine[]>('budget_editor_lines', () => [])
  const loading = useState<boolean>('budget_editor_loading', () => false)
  const saving = useState<boolean>('budget_editor_saving', () => false)

  const fetchLines = async (versionId: string) => {
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return []

    loading.value = true
    try {
      // Fetch version details which includes lines
      const version = await api<{ lines: BudgetLine[] }>(`/ledgers/${ledgerId}/budget/versions/${versionId}`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${accessToken.value}` }
      })
      lines.value = version.lines
      return version.lines
    } catch (err) {
      console.error(err)
      return []
    } finally {
      loading.value = false
    }
  }

  const updateLineApi = async (versionId: string, lineId: string, payload: BudgetLineUpdatePayload) => {
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return

    saving.value = true
    try {
      await api(`/ledgers/${ledgerId}/budget/versions/${versionId}/lines/${lineId}`, {
        method: 'PATCH',
        headers: { Authorization: `Bearer ${accessToken.value}` },
        body: payload
      })
    } finally {
      saving.value = false
    }
  }

  // Debounce updates to avoid flooding API
  const debouncedUpdate = debounce(updateLineApi, 500)

  const updateLine = (versionId: string, lineId: string, payload: BudgetLineUpdatePayload) => {
    // Optimistic update
    lines.value = lines.value.map((l) => (l.id === lineId ? { ...l, ...payload } : l))
    debouncedUpdate(versionId, lineId, payload)
  }

  const createLine = async (versionId: string, categoryId: string, percent: number) => {
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return

    const payload = { category_id: categoryId, percent, include_children: false }
    const created = await api<BudgetLine>(`/ledgers/${ledgerId}/budget/versions/${versionId}/lines`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${accessToken.value}` },
      body: payload
    })
    lines.value.push(created)
    return created
  }
  
  const deleteLine = async (versionId: string, lineId: string) => {
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return 
    
    await api(`/ledgers/${ledgerId}/budget/versions/${versionId}/lines/${lineId}`, {
      method: 'DELETE',
        headers: { Authorization: `Bearer ${accessToken.value}` }
    })
    lines.value = lines.value.filter(l => l.id !== lineId)
  }

  return {
    lines,
    loading,
    saving,
    fetchLines,
    updateLine,
    createLine,
    deleteLine
  }
}
