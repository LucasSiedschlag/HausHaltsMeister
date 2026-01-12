import { useApiClient } from '@shared/composables/useApiClient'
import { useAuth } from '#layers/auth/composables/useAuth'

export type CardNetwork = {
  code: string
  display_name: string
  is_active?: boolean
  created_at?: string
  updated_at?: string | null
}

type CardNetworkPayload = {
  code: string
  display_name: string
}

export const useCardNetworks = () => {
  const api = useApiClient()
  const { accessToken } = useAuth()

  const networks = useState<CardNetwork[]>('card_networks_list', () => [])
  const loading = useState<boolean>('card_networks_loading', () => false)
  const error = useState<string | null>('card_networks_error', () => null)

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const fetchNetworks = async () => {
    if (!(await ensureAccessToken())) return []
    loading.value = true
    error.value = null
    try {
      const payload = await api<CardNetwork[]>('/card-networks', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      networks.value = payload
      return payload
    } catch (err) {
      error.value = (err as Error)?.message || 'Erro ao carregar bandeiras'
      return []
    } finally {
      loading.value = false
    }
  }

  const createNetwork = async (payload: CardNetworkPayload) => {
    if (!(await ensureAccessToken())) return null
    const created = await api<CardNetwork>('/card-networks', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    networks.value = [created, ...networks.value]
    return created
  }

  const updateNetwork = async (code: string, payload: { display_name: string }) => {
    if (!(await ensureAccessToken())) return null
    const updated = await api<CardNetwork>(`/card-networks/${code}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      body: payload
    })
    networks.value = networks.value.map((item) => (item.code === updated.code ? updated : item))
    return updated
  }

  const deleteNetwork = async (code: string) => {
    if (!(await ensureAccessToken())) return false
    await api(`/card-networks/${code}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    networks.value = networks.value.filter((item) => item.code !== code)
    return true
  }

  return {
    networks,
    loading,
    error,
    fetchNetworks,
    createNetwork,
    updateNetwork,
    deleteNetwork
  }
}
