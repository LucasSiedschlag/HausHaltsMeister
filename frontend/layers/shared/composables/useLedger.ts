import { computed } from 'vue'
import { useApiClient } from '#layers/shared/composables/useApiClient'
import { useLedgerContext, type LedgerSummary } from '#layers/shared/composables/useLedgerContext'
import { useAuth } from '#layers/auth/composables/useAuth'

type LedgerCreatePayload = {
  name: string
  currency_code: string
}

export const useLedger = () => {
  const api = useApiClient()
  const ledgerContext = useLedgerContext()
  const ledgerCookie = useCookie<string | null>('hhm_ledger_id', { sameSite: 'lax' })
  const { accessToken, refresh } = useAuth()

  const selectLedger = async (ledgerId: string | null) => {
    if (!ledgerId) {
      ledgerContext.activeLedgerId.value = null
      ledgerCookie.value = null
      return
    }
    await ledgerContext.setActiveLedger(ledgerId)
    ledgerCookie.value = ledgerId
  }

  const currentLedgerId = computed({
    get: () => ledgerContext.activeLedgerId.value,
    set: (value: string | null) => {
      void selectLedger(value)
    }
  })
  const currentLedger = computed(() => ledgerContext.activeLedger.value)
  const ledgers = ledgerContext.ledgers
  const isLoading = ledgerContext.loading

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
    const payload = await ledgerContext.loadLedgers()
    if (!ledgerContext.activeLedgerId.value && ledgerCookie.value) {
      await ledgerContext.setActiveLedger(ledgerCookie.value)
    }
    return payload
  }

  const createLedger = async (payload: LedgerCreatePayload) => {
    if (!(await ensureAccessToken())) return null
    const created = await api<LedgerSummary>('/ledgers', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    ledgerContext.ledgers.value = [created, ...ledgerContext.ledgers.value]
    await ledgerContext.setActiveLedger(created.id)
    ledgerCookie.value = created.id
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
