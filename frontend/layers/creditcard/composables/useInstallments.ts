import { useApiClient } from '@shared/composables/useApiClient'
import { useAuth } from '#layers/auth/composables/useAuth'
import { createIdempotencyKey } from '@shared/utils/idempotency'

export type Installment = {
  id: string
  ledger_id: string
  plan_id: string
  installment_no: number
  due_month: string
  amount_cents: number
  status: string
  posted_transaction_id?: string | null
  paid_statement_id?: string | null
  created_at?: string
  updated_at?: string | null
}

export type PostingResult = {
  posted_count: number
  transaction_ids: string[]
}

export const useInstallments = () => {
  const api = useApiClient()
  const { accessToken } = useAuth()

  const installments = useState<Installment[]>('credit_card_installments', () => [])
  const loading = useState<boolean>('credit_card_installments_loading', () => false)
  const error = useState<string | null>('credit_card_installments_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchInstallments = async (cardId: string, filters?: { month?: string; status?: string }) => {
    if (!(await ensureAccessToken())) return []
    if (!cardId) return []
    loading.value = true
    error.value = null
    const params = new URLSearchParams()
    if (filters?.month) params.set('month', filters.month)
    if (filters?.status) params.set('status', filters.status)
    const query = params.toString()
    const path = query ? `/credit-cards/${cardId}/installments?${query}` : `/credit-cards/${cardId}/installments`
    try {
      const payload = await api<Installment[]>(path, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      installments.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar parcelas'
      return []
    } finally {
      loading.value = false
    }
  }

  const updateInstallment = async (cardId: string, installmentId: string, status: string) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const updated = await api<Installment>(`/credit-cards/${cardId}/installments/${installmentId}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: { status }
    })
    installments.value = installments.value.map((item) =>
      item.id === updated.id ? updated : item
    )
    return updated
  }

  const postMonth = async (cardId: string, month: string) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const params = new URLSearchParams({ month })
    return api<PostingResult>(`/credit-cards/${cardId}/post?${params.toString()}`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey()
    })
  }

  return {
    installments,
    loading,
    error,
    fetchInstallments,
    updateInstallment,
    postMonth
  }
}
