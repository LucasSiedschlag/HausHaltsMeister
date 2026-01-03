import { useAuth } from '#layers/auth/composables/useAuth'

export type LedgerRole = 'owner' | 'editor' | 'viewer'

export type Ledger = {
  id: string
  owner_user_id: string
  name: string
  currency_code: string
  created_at: string
  updated_at: string | null
}

export type LedgerMembership = Ledger & {
  role?: LedgerRole
}

type LedgerCreatePayload = {
  name: string
  currency_code: string
}

export const useLedger = () => {
  const ledgerCookie = useCookie<string | null>('hhm_ledger_id', { sameSite: 'lax' })
  const ledgers = useState<LedgerMembership[]>('ledgers', () => [])
  const isLoading = useState<boolean>('ledgers_loading', () => false)
  const api = useApiClient()
  const { accessToken, refresh } = useAuth()

  const currentLedgerId = computed({
    get: () => ledgerCookie.value,
    set: (value: string | null) => {
      ledgerCookie.value = value
    }
  })

  const currentLedger = computed(() => {
    if (!currentLedgerId.value) return null
    return ledgers.value.find((item) => item.id === currentLedgerId.value) || null
  })

  const ensureAccessToken = async () => {
    if (accessToken.value) return true
    try {
      await refresh()
      return Boolean(accessToken.value)
    } catch {
      return false
    }
  }

  const fetchLedgers = async () => {
    if (!(await ensureAccessToken())) return []
    isLoading.value = true
    try {
      const payload = await api<LedgerMembership[]>('/ledgers', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      ledgers.value = payload
      return payload
    } catch {
      return ledgers.value
    } finally {
      isLoading.value = false
    }
  }

  const selectLedger = (ledgerId: string | null) => {
    currentLedgerId.value = ledgerId
  }

  const createLedger = async (payload: LedgerCreatePayload) => {
    if (!(await ensureAccessToken())) return null
    const created = await api<LedgerMembership>('/ledgers', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    const createdEntry = created.role ? created : { ...created, role: 'owner' as LedgerRole }
    ledgers.value = [createdEntry, ...ledgers.value]
    currentLedgerId.value = createdEntry.id
    return created
  }

  return {
    ledgers,
    isLoading,
    currentLedgerId,
    currentLedger,
    fetchLedgers,
    selectLedger,
    createLedger
  }
}
