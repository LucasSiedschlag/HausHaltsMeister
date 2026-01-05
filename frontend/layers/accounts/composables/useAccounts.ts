import { useApiClient } from '@shared/composables/useApiClient'
import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'

export type AccountType = 'cash' | 'investment' | 'credit_card'

export type Account = {
  id: string
  ledger_id: string
  name: string
  type: AccountType
  is_active: boolean
  created_at?: string
  updated_at?: string | null
}

type AccountCreatePayload = {
  name: string
  type: AccountType
  is_active: boolean
}

type AccountUpdatePayload = {
  name: string
  is_active: boolean
}

export const useAccounts = () => {
  const api = useApiClient()
  const { currentLedgerId } = useLedger()
  const { accessToken } = useAuth()

  const accounts = useState<Account[]>('accounts_list', () => [])
  const loading = useState<boolean>('accounts_loading', () => false)
  const error = useState<string | null>('accounts_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchAccounts = async () => {
    if (!(await ensureAccessToken())) return []
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return []
    loading.value = true
    error.value = null
    try {
      const payload = await api<Account[]>(`/ledgers/${ledgerId}/accounts`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      accounts.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar contas'
      return []
    } finally {
      loading.value = false
    }
  }

  const createAccount = async (payload: AccountCreatePayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    const created = await api<Account>(`/ledgers/${ledgerId}/accounts`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    accounts.value = [created, ...accounts.value]
    return created
  }

  const updateAccount = async (accountId: string, payload: AccountUpdatePayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    const updated = await api<Account>(`/ledgers/${ledgerId}/accounts/${accountId}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    accounts.value = accounts.value.map((item) => (item.id === updated.id ? updated : item))
    return updated
  }

  const deactivateAccount = async (accountId: string) => {
    if (!(await ensureAccessToken())) return false
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return false
    await api(`/ledgers/${ledgerId}/accounts/${accountId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    accounts.value = accounts.value.map((item) =>
      item.id === accountId ? { ...item, is_active: false } : item
    )
    return true
  }

  return {
    accounts,
    loading,
    error,
    fetchAccounts,
    createAccount,
    updateAccount,
    deactivateAccount
  }
}
