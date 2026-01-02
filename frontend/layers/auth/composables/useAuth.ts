import { useAnalytics } from '@shared/composables/useAnalytics'

type AuthUser = {
  id: string
  email: string
  display_name: string
  avatar_url?: string | null
  email_verified_at?: string | null
}

type AuthResponse = {
  access_token: string
  token_type: string
  expires_in: number
  user: AuthUser
}

type AuthSession = {
  id: string
  created_at: string
  expires_at: string
  user_agent?: string
  ip?: string
  device_name?: string
  is_current?: boolean
}

type ApiError = {
  code?: string
  message?: string
  details?: Record<string, string>
  error?: {
    code?: string
    message?: string
    details?: Record<string, string>
  }
}

const extractErrorMessage = (err: unknown) => {
  const data = (err as { data?: ApiError })?.data
  return data?.message || data?.error?.message || 'Erro inesperado'
}

const extractErrorCode = (err: unknown) => {
  const data = (err as { data?: ApiError })?.data
  return data?.code || data?.error?.code || ''
}

const extractErrorDetails = (err: unknown) => {
  const data = (err as { data?: ApiError })?.data
  return data?.details || data?.error?.details || {}
}

export const useAuth = () => {
  const accessToken = useState<string | null>('access_token', () => null)
  const user = useState<AuthUser | null>('auth_user', () => null)
  const expiryTimer = useState<number | null>('auth_expiry_timer', () => null)
  const expiryWatcherSet = useState<boolean>('auth_expiry_watcher', () => false)
  const expiryInterval = useState<number | null>('auth_expiry_interval', () => null)
  const sessionExpired = useState<boolean>('auth_session_expired', () => false)
  const api = useApiClient()
  const { track } = useAnalytics()
  const preferencesNeeded = useState<boolean>('preferences_needed', () => false)
  const router = useRouter()

  const clearExpiryTimer = () => {
    if (expiryTimer.value !== null && process.client) {
      window.clearTimeout(expiryTimer.value)
    }
    expiryTimer.value = null
  }

  const clearExpiryWatcher = () => {
    if (!process.client) return
    if (expiryInterval.value !== null) {
      window.clearInterval(expiryInterval.value)
    }
    expiryInterval.value = null
    if (expiryWatcherSet.value) {
      window.removeEventListener('focus', checkExpiry)
      document.removeEventListener('visibilitychange', checkExpiry)
    }
    expiryWatcherSet.value = false
  }

  const parseAccessTokenExp = (token: string) => {
    try {
      const [, payload] = token.split('.')
      if (!payload) return null
      const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
      const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=')
      const decoded = JSON.parse(atob(padded)) as { exp?: number }
      return decoded.exp ?? null
    } catch {
      return null
    }
  }

  const handleSessionExpired = async () => {
    clearExpiryTimer()
    clearExpiryWatcher()
    accessToken.value = null
    user.value = null
    sessionExpired.value = true
  }

  const acknowledgeSessionExpired = async () => {
    sessionExpired.value = false
    if (process.client) {
      await router.push('/auth/login')
    }
  }

  const checkExpiry = () => {
    if (!process.client || !accessToken.value) return
    const exp = parseAccessTokenExp(accessToken.value)
    if (!exp) return
    if (Date.now() >= exp * 1000) {
      void handleSessionExpired()
    }
  }

  const scheduleExpiry = (token: string) => {
    if (!process.client) return
    clearExpiryTimer()
    const exp = parseAccessTokenExp(token)
    if (!exp) return
    const delay = exp * 1000 - Date.now()
    if (delay <= 0) {
      void handleSessionExpired()
      return
    }
    expiryTimer.value = window.setTimeout(() => {
      void handleSessionExpired()
    }, delay)
  }

  const ensureExpiryWatcher = () => {
    if (!process.client || expiryWatcherSet.value) return
    expiryWatcherSet.value = true
    expiryInterval.value = window.setInterval(checkExpiry, 30000)
    window.addEventListener('focus', checkExpiry)
    document.addEventListener('visibilitychange', checkExpiry)
    checkExpiry()
  }

  const setSession = (payload: AuthResponse) => {
    accessToken.value = payload.access_token
    user.value = payload.user
    scheduleExpiry(payload.access_token)
    ensureExpiryWatcher()
    preferencesNeeded.value = true
  }

  const login = async (email: string, password: string, remember: boolean) => {
    const payload = await api<AuthResponse>('/auth/login', {
      method: 'POST',
      body: { email, password, remember }
    })
    setSession(payload)
    track('auth.login', { method: 'password', remember })
    return payload
  }

  const signup = async (email: string, password: string, displayName: string) => {
    const payload = await api<AuthResponse>('/auth/signup', {
      method: 'POST',
      body: { email, password, display_name: displayName }
    })
    setSession(payload)
    track('auth.signup', { method: 'password' })
    return payload
  }

  const refresh = async () => {
    const payload = await api<AuthResponse>('/auth/refresh', { method: 'POST' })
    setSession(payload)
    track('auth.refresh', {})
    return payload
  }

  const startOAuth = (provider: 'google' | 'github') => {
    const config = useRuntimeConfig()
    if (process.server) return
    const redirectURI = `${window.location.origin}/auth/oauth/callback`
    track('auth.oauth.start', { provider })
    const oauthBase = config.public.oauthBase || config.public.apiBase
    window.location.href = `${oauthBase}/auth/oauth/${provider}/start?redirect_uri=${encodeURIComponent(
      redirectURI
    )}`
  }

  const me = async () => {
    if (!accessToken.value) return null
    const payload = await api<AuthUser>('/auth/me', {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    user.value = payload
    return payload
  }

  const logout = async () => {
    await api('/auth/logout', { method: 'POST' })
    accessToken.value = null
    user.value = null
    clearExpiryTimer()
    clearExpiryWatcher()
  }

  const clearSession = () => {
    accessToken.value = null
    user.value = null
    clearExpiryTimer()
    clearExpiryWatcher()
  }

  const listSessions = async () => {
    if (!accessToken.value) {
      await refresh()
    }
    if (!accessToken.value) return []
    return api<AuthSession[]>('/auth/sessions', {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
  }

  const revokeSession = async (sessionId: string) => {
    if (!accessToken.value) {
      await refresh()
    }
    if (!accessToken.value) return
    await api(`/auth/sessions/${sessionId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
  }

  const logoutAll = async () => {
    if (!accessToken.value) {
      await refresh()
    }
    if (!accessToken.value) return
    await api('/auth/logout-all', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
  }

  const forgotPassword = async (email: string) => {
    await api('/auth/forgot-password', {
      method: 'POST',
      body: { email }
    })
  }

  return {
    accessToken,
    user,
    login,
    signup,
    refresh,
    me,
    logout,
    clearSession,
    sessionExpired,
    acknowledgeSessionExpired,
    listSessions,
    revokeSession,
    logoutAll,
    forgotPassword,
    startOAuth,
    extractErrorMessage,
    extractErrorCode,
    extractErrorDetails
  }
}
