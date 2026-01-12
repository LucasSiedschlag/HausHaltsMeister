import { useApiClient } from '@shared/composables/useApiClient'
import { useAuth } from '#layers/auth/composables/useAuth'
import { createIdempotencyKey } from '@shared/utils/idempotency'

export type InstallmentPlan = {
  id: string
  ledger_id: string
  credit_card_id: string
  purchase_occurred_at: string
  merchant?: string | null
  description: string
  category_id: string
  total_amount_cents: number
  installments_count: number
  installment_amount_cents: number
  first_due_month: string
  status: string
  created_by_user_id?: string
  created_at?: string
  updated_at?: string | null
}

type InstallmentPlanPayload = {
  purchase_occurred_at: string
  merchant?: string | null
  description: string
  category_id: string
  total_amount_cents: number
  installments_count: number
  installment_amount_cents: number
  first_due_month: string
}

export const useInstallmentPlans = () => {
  const api = useApiClient()
  const { accessToken } = useAuth()

  const plans = useState<InstallmentPlan[]>('credit_card_plans', () => [])
  const loading = useState<boolean>('credit_card_plans_loading', () => false)
  const error = useState<string | null>('credit_card_plans_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchPlans = async (cardId: string, status?: string) => {
    if (!(await ensureAccessToken())) return []
    if (!cardId) return []
    loading.value = true
    error.value = null
    const params = new URLSearchParams()
    if (status) params.set('status', status)
    const query = params.toString()
    const path = query ? `/credit-cards/${cardId}/plans?${query}` : `/credit-cards/${cardId}/plans`
    try {
      const payload = await api<InstallmentPlan[]>(path, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      plans.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar planos'
      return []
    } finally {
      loading.value = false
    }
  }

  const createPlan = async (cardId: string, payload: InstallmentPlanPayload) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const created = await api<InstallmentPlan>(`/credit-cards/${cardId}/plans`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey(),
      body: payload
    })
    plans.value = [created, ...plans.value]
    return created
  }

  const cancelPlan = async (cardId: string, planId: string) => {
    if (!(await ensureAccessToken())) return false
    if (!cardId) return false
    await api(`/credit-cards/${cardId}/plans/${planId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    plans.value = plans.value.map((plan) =>
      plan.id === planId ? { ...plan, status: 'cancelled' } : plan
    )
    return true
  }

  return {
    plans,
    loading,
    error,
    fetchPlans,
    createPlan,
    cancelPlan
  }
}
