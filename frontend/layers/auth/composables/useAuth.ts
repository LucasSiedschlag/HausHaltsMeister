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
  const api = useApiClient()
  const { track } = useAnalytics()

  const setSession = (payload: AuthResponse) => {
    accessToken.value = payload.access_token
    user.value = payload.user
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
    forgotPassword,
    startOAuth,
    extractErrorMessage,
    extractErrorCode,
    extractErrorDetails
  }
}
