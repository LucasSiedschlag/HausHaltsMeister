import { useApiClient } from '@shared/composables/useApiClient'
import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'

export type BalanceItem = {
  account_id: string
  account_name: string
  account_type: string
  balance_cents: number
}

export type BalancesReport = {
  ledger_id: string
  month: string
  items: BalanceItem[]
}

export const useBalancesReport = () => {
  const api = useApiClient()
  const { currentLedgerId } = useLedger()
  const { accessToken } = useAuth()

  const balances = useState<BalancesReport | null>('report_balances', () => null)
  const loading = useState<boolean>('report_balances_loading', () => false)
  const error = useState<string | null>('report_balances_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchBalances = async (month: string) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    loading.value = true
    error.value = null
    const params = new URLSearchParams({ month })
    const path = `/ledgers/${ledgerId}/reports/balances?${params.toString()}`
    try {
      const payload = await api<BalancesReport>(path, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      balances.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar saldos'
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    balances,
    loading,
    error,
    fetchBalances
  }
}
