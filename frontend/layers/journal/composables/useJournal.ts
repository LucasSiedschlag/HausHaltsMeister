import { useApiClient } from '@shared/composables/useApiClient'
import { useLedger } from '@shared/composables/useLedger'
import { useAuth } from '#layers/auth/composables/useAuth'
import { createIdempotencyKey } from '@shared/utils/idempotency'

export type EntryKind = 'normal' | 'transfer' | 'adjust'

export type TransactionEntry = {
  id: string
  account_id: string
  category_id?: string | null
  kind: EntryKind
  amount_cents: number
  memo?: string | null
}

export type Transaction = {
  id: string
  ledger_id: string
  occurred_at: string
  description: string
  notes?: string | null
  created_at?: string
  updated_at?: string | null
  entries: TransactionEntry[]
}

export type TransactionCursor = {
  cursor_occurred_at: string
  cursor_id: string
}

export type TransactionListResponse = {
  items: Transaction[]
  next_cursor?: TransactionCursor | null
}

export type TransactionEntryInput = {
  account_id: string
  category_id?: string | null
  kind: EntryKind
  amount_cents: number
  memo?: string | null
}

type TransactionCreatePayload = {
  occurred_at: string
  description: string
  notes?: string | null
  entries: TransactionEntryInput[]
}

type TransactionUpdatePayload = {
  occurred_at?: string
  description?: string
  notes?: string | null
  entries?: TransactionEntryInput[]
}

export type JournalFilters = {
  from?: string
  to?: string
  account_id?: string
  category_id?: string
  q?: string
  limit?: number
  cursor_occurred_at?: string
  cursor_id?: string
}

export const useJournal = () => {
  const api = useApiClient()
  const { currentLedgerId } = useLedger()
  const { accessToken } = useAuth()

  const transactions = useState<Transaction[]>('journal_transactions', () => [])
  const nextCursor = useState<TransactionCursor | null>('journal_next_cursor', () => null)
  const lastPageSize = useState<number>('journal_last_page_size', () => 0)
  const loading = useState<boolean>('journal_loading', () => false)
  const error = useState<string | null>('journal_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const buildQuery = (filters?: JournalFilters) => {
    const params = new URLSearchParams()
    if (!filters) return params
    if (filters.from) params.set('from', filters.from)
    if (filters.to) params.set('to', filters.to)
    if (filters.account_id) params.set('account_id', filters.account_id)
    if (filters.category_id) params.set('category_id', filters.category_id)
    if (filters.q) params.set('q', filters.q)
    if (filters.limit) params.set('limit', String(filters.limit))
    if (filters.cursor_occurred_at) params.set('cursor_occurred_at', filters.cursor_occurred_at)
    if (filters.cursor_id) params.set('cursor_id', filters.cursor_id)
    return params
  }

  const fetchTransactions = async (filters?: JournalFilters, options?: { append?: boolean }) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    loading.value = true
    error.value = null
    const params = buildQuery(filters)
    const query = params.toString()
    const path = query ? `/ledgers/${ledgerId}/transactions?${query}` : `/ledgers/${ledgerId}/transactions`
    try {
      const payload = await api<TransactionListResponse>(path, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      if (options?.append) {
        transactions.value = [...transactions.value, ...payload.items]
      } else {
        transactions.value = payload.items
      }
      nextCursor.value = payload.next_cursor ?? null
      lastPageSize.value = payload.items.length
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar transacoes'
      return null
    } finally {
      loading.value = false
    }
  }

  const createTransaction = async (payload: TransactionCreatePayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    const created = await api<Transaction>(`/ledgers/${ledgerId}/transactions`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey(),
      body: payload
    })
    transactions.value = [created, ...transactions.value]
    return created
  }

  const updateTransaction = async (transactionId: string, payload: TransactionUpdatePayload) => {
    if (!(await ensureAccessToken())) return null
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return null
    const updated = await api<Transaction>(`/ledgers/${ledgerId}/transactions/${transactionId}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    transactions.value = transactions.value.map((item) => (item.id === updated.id ? updated : item))
    return updated
  }

  const deleteTransaction = async (transactionId: string) => {
    if (!(await ensureAccessToken())) return false
    const ledgerId = currentLedgerId.value
    if (!ledgerId) return false
    await api(`/ledgers/${ledgerId}/transactions/${transactionId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    transactions.value = transactions.value.filter((item) => item.id !== transactionId)
    return true
  }

  return {
    transactions,
    nextCursor,
    lastPageSize,
    loading,
    error,
    fetchTransactions,
    createTransaction,
    updateTransaction,
    deleteTransaction
  }
}
