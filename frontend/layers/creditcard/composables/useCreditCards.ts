import { useApiClient } from '@shared/composables/useApiClient'
import { useAuth } from '#layers/auth/composables/useAuth'

export type CreditCard = {
  id: string
  ledger_id: string
  parent_account_id: string
  liability_account_id: string
  label?: string | null
  brand: string
  last4?: string | null
  cvv?: string | null
  holder_name?: string | null
  active: boolean
  color?: string | null
  style?: string | null
  closing_day: number
  due_day: number
  created_at?: string
  updated_at?: string | null
}

type CreditCardPayload = {
  label?: string | null
  brand: string
  last4?: string | null
  cvv?: string | null
  holder_name?: string | null
  active?: boolean
  color?: string | null
  style?: string | null
  closing_day: number
  due_day: number
}

export const useCreditCards = () => {
  const api = useApiClient()
  const { accessToken } = useAuth()

  const cards = useState<CreditCard[]>('credit_cards_list', () => [])
  const loading = useState<boolean>('credit_cards_loading', () => false)
  const error = useState<string | null>('credit_cards_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchCards = async (accountId: string) => {
    if (!(await ensureAccessToken())) return []
    if (!accountId) return []
    loading.value = true
    error.value = null
    try {
      const payload = await api<CreditCard[]>(`/accounts/${accountId}/credit-cards`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      cards.value = [
        ...cards.value.filter((card) => card.parent_account_id !== accountId),
        ...payload
      ]
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar cartões'
      return []
    } finally {
      loading.value = false
    }
  }

  const fetchCard = async (cardId: string) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const payload = await api<CreditCard>(`/credit-cards/${cardId}`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    cards.value = cards.value.some((item) => item.id === payload.id)
      ? cards.value.map((item) => (item.id === payload.id ? payload : item))
      : [payload, ...cards.value]
    return payload
  }

  const createCard = async (accountId: string, payload: CreditCardPayload) => {
    if (!(await ensureAccessToken())) return null
    if (!accountId) return null
    const created = await api<CreditCard>(`/accounts/${accountId}/credit-cards`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    cards.value = [created, ...cards.value]
    return created
  }

  const updateCard = async (cardId: string, payload: CreditCardPayload) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const updated = await api<CreditCard>(`/credit-cards/${cardId}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    cards.value = cards.value.map((item) => (item.id === updated.id ? updated : item))
    return updated
  }

  const deleteCard = async (cardId: string) => {
    if (!(await ensureAccessToken())) return false
    if (!cardId) return false
    await api(`/credit-cards/${cardId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    cards.value = cards.value.filter((item) => item.id !== cardId)
    return true
  }

  return {
    cards,
    loading,
    error,
    fetchCards,
    fetchCard,
    createCard,
    updateCard,
    deleteCard
  }
}
