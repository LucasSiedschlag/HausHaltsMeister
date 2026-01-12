import { useApiClient } from '@shared/composables/useApiClient'
import { useAuth } from '#layers/auth/composables/useAuth'
import { createIdempotencyKey } from '@shared/utils/idempotency'

export type Statement = {
  id: string
  ledger_id: string
  credit_card_id: string
  statement_month: string
  closing_date: string
  due_date: string
  total_charges_cents: number
  total_payments_cents: number
  status: string
  payment_transaction_id?: string | null
  created_at?: string
  updated_at?: string | null
}

export type StatementPayPayload = {
  statement_id: string
  payment_date: string
  pay_amount_cents: number
  paying_account_id: string
}

export const useStatements = () => {
  const api = useApiClient()
  const { accessToken } = useAuth()

  const statements = useState<Statement[]>('credit_card_statements', () => [])
  const loading = useState<boolean>('credit_card_statements_loading', () => false)
  const error = useState<string | null>('credit_card_statements_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchStatements = async (cardId: string, month?: string) => {
    if (!(await ensureAccessToken())) return []
    if (!cardId) return []
    loading.value = true
    error.value = null
    const params = new URLSearchParams()
    if (month) params.set('month', month)
    const query = params.toString()
    const path = query ? `/credit-cards/${cardId}/statements?${query}` : `/credit-cards/${cardId}/statements`
    try {
      const payload = await api<Statement[]>(path, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      statements.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar faturas'
      return []
    } finally {
      loading.value = false
    }
  }

  const fetchStatement = async (cardId: string, statementId: string) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const payload = await api<Statement>(`/credit-cards/${cardId}/statements/${statementId}`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    statements.value = statements.value.some((item) => item.id === payload.id)
      ? statements.value.map((item) => (item.id === payload.id ? payload : item))
      : [payload, ...statements.value]
    return payload
  }

  const closeStatement = async (cardId: string, month: string) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const params = new URLSearchParams({ month })
    const updated = await api<Statement>(`/credit-cards/${cardId}/statements/close?${params.toString()}`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey()
    })
    statements.value = statements.value.some((item) => item.id === updated.id)
      ? statements.value.map((item) => (item.id === updated.id ? updated : item))
      : [updated, ...statements.value]
    return updated
  }

  const payStatement = async (cardId: string, payload: StatementPayPayload) => {
    if (!(await ensureAccessToken())) return null
    if (!cardId) return null
    const updated = await api<Statement>(`/credit-cards/${cardId}/statements/pay`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey(),
      body: payload
    })
    statements.value = statements.value.some((item) => item.id === updated.id)
      ? statements.value.map((item) => (item.id === updated.id ? updated : item))
      : [updated, ...statements.value]
    return updated
  }

  return {
    statements,
    loading,
    error,
    fetchStatements,
    fetchStatement,
    closeStatement,
    payStatement
  }
}
