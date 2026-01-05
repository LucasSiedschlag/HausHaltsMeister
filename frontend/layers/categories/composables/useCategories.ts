import { useApiClient } from '@shared/composables/useApiClient'
import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'

export type CategoryDirection = 'in' | 'out'

export type Category = {
  id: string
  ledger_id: string
  parent_id?: string | null
  name: string
  direction: CategoryDirection
  is_budget_base: boolean
  is_budget_relevant: boolean
  is_active: boolean
  created_at?: string
  updated_at?: string | null
}

type CategoryCreatePayload = {
  parent_id?: string | null
  name: string
  direction: CategoryDirection
  is_budget_base: boolean
  is_budget_relevant: boolean
}

type CategoryUpdatePayload = {
  parent_id?: string | null
  name: string
  is_budget_base: boolean
  is_budget_relevant: boolean
  is_active: boolean
}

export const useCategories = () => {
  const api = useApiClient()
  const { currentLedgerId } = useLedger()
  const { accessToken } = useAuth()

  const categories = useState<Category[]>('categories_list', () => [])
  const loading = useState<boolean>('categories_loading', () => false)
  const error = useState<string | null>('categories_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchCategories = async (filters?: { direction?: CategoryDirection | 'all'; active?: boolean | 'all' }) => {
    if (!(await ensureAccessToken())) return []
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return []
    loading.value = true
    error.value = null
    const params = new URLSearchParams()
    if (filters?.direction && filters.direction !== 'all') {
      params.set('direction', filters.direction)
    }
    if (filters?.active !== undefined && filters.active !== 'all') {
      params.set('active', String(filters.active))
    }
    const query = params.toString()
    const path = query ? `/ledgers/${ledgerId}/categories?${query}` : `/ledgers/${ledgerId}/categories`
    try {
      const payload = await api<Category[]>(path, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      categories.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar categorias'
      return []
    } finally {
      loading.value = false
    }
  }

  const createCategory = async (payload: CategoryCreatePayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    const created = await api<Category>(`/ledgers/${ledgerId}/categories`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    categories.value = [created, ...categories.value]
    return created
  }

  const updateCategory = async (categoryId: string, payload: CategoryUpdatePayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    const updated = await api<Category>(`/ledgers/${ledgerId}/categories/${categoryId}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    categories.value = categories.value.map((item) => (item.id === updated.id ? updated : item))
    return updated
  }

  const deactivateCategory = async (categoryId: string) => {
    if (!(await ensureAccessToken())) return false
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return false
    await api(`/ledgers/${ledgerId}/categories/${categoryId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    categories.value = categories.value.map((item) =>
      item.id === categoryId ? { ...item, is_active: false } : item
    )
    return true
  }

  return {
    categories,
    loading,
    error,
    fetchCategories,
    createCategory,
    updateCategory,
    deactivateCategory
  }
}
