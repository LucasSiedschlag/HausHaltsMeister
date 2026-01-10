import { useApiClient } from '@shared/composables/useApiClient'
import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'
import { createIdempotencyKey } from '@shared/utils/idempotency'
import type { Transaction } from '#layers/journal/composables/useJournal'

export type InvestmentSummary = {
  ledger_id: string
  from: string
  to: string
  total_contributions: number
  total_redemptions: number
  total_earnings: number
  total_losses: number
  net_variation: number
}

type InvestmentPayload = {
  amount_cents: number
  occurred_at: string
  memo?: string | null
}

export const useInvestments = () => {
  const api = useApiClient()
  const { currentLedgerId } = useLedger()
  const { accessToken } = useAuth()

  const summary = useState<InvestmentSummary | null>('investments_summary', () => null)
  const loading = useState<boolean>('investments_summary_loading', () => false)
  const error = useState<string | null>('investments_summary_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchSummary = async (from: string, to: string) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    loading.value = true
    error.value = null
    const params = new URLSearchParams({ from, to })
    const path = `/ledgers/${ledgerId}/investments/summary?${params.toString()}`
    try {
      const payload = await api<InvestmentSummary>(path, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      summary.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar resumo de investimentos'
      return null
    } finally {
      loading.value = false
    }
  }

  const contribute = async (payload: InvestmentPayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    return api<Transaction>(`/ledgers/${ledgerId}/investments/contributions`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey(),
      body: payload
    })
  }

  const redeem = async (payload: InvestmentPayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    return api<Transaction>(`/ledgers/${ledgerId}/investments/redemptions`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey(),
      body: payload
    })
  }

  const earn = async (payload: InvestmentPayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    return api<Transaction>(`/ledgers/${ledgerId}/investments/earnings`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey(),
      body: payload
    })
  }

  return {
    summary,
    loading,
    error,
    fetchSummary,
    contribute,
    redeem,
    earn
  }
}
