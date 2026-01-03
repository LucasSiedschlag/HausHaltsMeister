import { useAuth } from '#layers/auth/composables/useAuth'

type ThemeMode = 'light' | 'dark' | 'system'
type ThemePalette =
  | 'default'
  | 'red'
  | 'rose'
  | 'orange'
  | 'green'
  | 'yellow'
  | 'violet'
  | 'monochrome'
type ThemeTone = 'vivid' | 'pastel' | 'muted'
type Locale = 'pt-BR' | 'en-US'
type CompactMode = 'comfortable' | 'compact' | 'dense'
type FontScale = 'sm' | 'md' | 'lg'

export type UserPreferences = {
  user_id: string
  theme_mode: ThemeMode
  theme_palette: ThemePalette
  theme_tone: ThemeTone
  locale: Locale
  compact_mode: CompactMode
  font_scale: FontScale
  notify_card_close: boolean
  notify_budget_over: boolean
  notify_payables: boolean
  created_at: string
  updated_at?: string | null
}

type PreferencesUpdate = Partial<
  Pick<
    UserPreferences,
    | 'theme_mode'
    | 'theme_palette'
    | 'theme_tone'
    | 'locale'
    | 'compact_mode'
    | 'font_scale'
    | 'notify_card_close'
    | 'notify_budget_over'
    | 'notify_payables'
  >
>

const applyThemeAttributes = (prefs: UserPreferences) => {
  if (!process.client) return
  const root = document.documentElement
  root.dataset.themePalette = prefs.theme_palette
  root.dataset.themeTone = prefs.theme_tone
}

export const usePreferences = () => {
  const preferences = useState<UserPreferences | null>('user_preferences', () => null)
  const isLoading = useState<boolean>('preferences_loading', () => false)
  const isSaving = useState<boolean>('preferences_saving', () => false)
  const preferencesNeeded = useState<boolean>('preferences_needed', () => false)
  const route = useRoute()
  const api = useApiClient()
  const { accessToken, refresh } = useAuth()
  const colorMode = useColorMode()
  const { locale, setLocale } = useI18n()

  const ensureAccessToken = async () => {
    if (accessToken.value) return true
    try {
      await refresh()
      return Boolean(accessToken.value)
    } catch {
      return false
    }
  }

  const applyPreferences = (prefs: UserPreferences) => {
    const localeMatch = route.path.match(/^\/([^/]+)(?:\/|$)/)
    const routeLocale =
      localeMatch && (localeMatch[1] === 'pt-BR' || localeMatch[1] === 'en-US')
        ? localeMatch[1]
        : null
    colorMode.preference = prefs.theme_mode
    if (!routeLocale && locale.value !== prefs.locale) {
      void setLocale(prefs.locale)
    } else if (routeLocale === prefs.locale && locale.value !== routeLocale) {
      void setLocale(routeLocale)
    }
    applyThemeAttributes(prefs)
  }

  const fetchPreferences = async () => {
    if (!(await ensureAccessToken())) return null
    isLoading.value = true
    try {
      const payload = await api<UserPreferences>('/me/preferences', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        }
      })
      preferences.value = payload
      applyPreferences(payload)
      return payload
    } finally {
      isLoading.value = false
    }
  }

  const updatePreferences = async (payload: PreferencesUpdate) => {
    if (!(await ensureAccessToken())) return null
    isSaving.value = true
    try {
      const updated = await api<UserPreferences>('/me/preferences', {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${accessToken.value}`
        },
        body: payload
      })
      preferences.value = updated
      applyPreferences(updated)
      return updated
    } finally {
      isSaving.value = false
    }
  }

  watch(
    () => preferences.value,
    (value) => {
      if (value) {
        applyPreferences(value)
      }
    },
    { immediate: true }
  )

  watch(
    () => preferencesNeeded.value,
    async (needed) => {
      if (!needed || preferences.value) return
      preferencesNeeded.value = false
      await fetchPreferences()
    }
  )

  watch(
    () => colorMode.value,
    () => {
      if (preferences.value) {
        applyPreferences(preferences.value)
      }
    }
  )

  return {
    preferences,
    isLoading,
    isSaving,
    fetchPreferences,
    updatePreferences
  }
}
