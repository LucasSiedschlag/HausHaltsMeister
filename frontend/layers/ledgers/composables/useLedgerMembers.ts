import { useApiClient } from '@shared/composables/useApiClient'
import { useAuth } from '#layers/auth/composables/useAuth'
import type { LedgerRole } from '#layers/shared/utils/ledger-roles'

export type LedgerMember = {
  ledger_id: string
  user_id: string
  role: LedgerRole
  display_name: string
  email: string
  avatar_url?: string | null
  created_at: string
  updated_at?: string | null
}

export const useLedgerMembers = () => {
  const api = useApiClient()
  const { accessToken } = useAuth()

  const membersByLedger = useState<Record<string, LedgerMember[]>>('ledger_members_by_ledger', () => ({}))
  const loadingByLedger = useState<Record<string, boolean>>('ledger_members_loading', () => ({}))
  const errorByLedger = useState<Record<string, string | null>>('ledger_members_error', () => ({}))

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchMembers = async (ledgerId: string) => {
    if (!(await ensureAccessToken())) return []
    loadingByLedger.value[ledgerId] = true
    errorByLedger.value[ledgerId] = null
    try {
      const payload = await api<LedgerMember[]>(`/ledgers/${ledgerId}/members`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      membersByLedger.value[ledgerId] = payload
      return payload
    } catch (err) {
      errorByLedger.value[ledgerId] = (err as Error)?.message || 'Erro ao carregar membros'
      return []
    } finally {
      loadingByLedger.value[ledgerId] = false
    }
  }

  const inviteMember = async (ledgerId: string, email: string, role: LedgerRole) => {
    if (!(await ensureAccessToken())) return null
    const created = await api<LedgerMember>(`/ledgers/${ledgerId}/members`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: {
        email,
        role
      }
    })
    membersByLedger.value[ledgerId] = [...(membersByLedger.value[ledgerId] ?? []), created]
    return created
  }

  const updateRole = async (ledgerId: string, memberId: string, role: LedgerRole) => {
    if (!(await ensureAccessToken())) return null
    const updated = await api<LedgerMember>(`/ledgers/${ledgerId}/members/${memberId}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: {
        role
      }
    })
    membersByLedger.value[ledgerId] = (membersByLedger.value[ledgerId] ?? []).map((member) =>
      member.user_id === memberId ? updated : member
    )
    return updated
  }

  const removeMember = async (ledgerId: string, memberId: string) => {
    if (!(await ensureAccessToken())) return false
    await api(`/ledgers/${ledgerId}/members/${memberId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    membersByLedger.value[ledgerId] = (membersByLedger.value[ledgerId] ?? []).filter(
      (member) => member.user_id !== memberId
    )
    return true
  }

  return {
    membersByLedger,
    loadingByLedger,
    errorByLedger,
    fetchMembers,
    inviteMember,
    updateRole,
    removeMember
  }
}
