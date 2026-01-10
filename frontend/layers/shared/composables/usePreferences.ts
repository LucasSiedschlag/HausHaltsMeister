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
  default_ledger_id: string | null
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
    | 'default_ledger_id'
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
  const api = useApiClient()
  const { accessToken } = useAuth()
  const colorMode = useColorMode()
  const nuxtApp = useNuxtApp()
  const i18n = nuxtApp.$i18n

  const getLocale = () => {
    if (!i18n) return null
    const localeValue = (i18n.locale as { value?: string } | string | undefined)
    if (typeof localeValue === 'string') {
      return localeValue
    }
    return localeValue?.value ?? null
  }

  const setLocaleSafe = async (value: Locale) => {
    if (!i18n) return
    if (typeof i18n.setLocale === 'function') {
      await i18n.setLocale(value)
      return
    }
    const localeValue = i18n.locale as { value?: string } | string | undefined
    if (localeValue && typeof localeValue === 'object' && 'value' in localeValue) {
      localeValue.value = value
    }
  }

  const ensureAccessToken = async () => {
    return Boolean(accessToken.value)
  }

  const applyPreferences = (prefs: UserPreferences) => {
    // Get route inside function to avoid issues with middleware context
    const route = useRoute()
    const localeMatch = route.path.match(/^\/([^/]+)(?:\/|$)/)
    const routeLocale =
      localeMatch && (localeMatch[1] === 'pt-BR' || localeMatch[1] === 'en-US')
        ? localeMatch[1]
        : null
    colorMode.preference = prefs.theme_mode
    const currentLocale = getLocale()
    if (!routeLocale && currentLocale && currentLocale !== prefs.locale) {
      void setLocaleSafe(prefs.locale)
    } else if (routeLocale === prefs.locale && currentLocale !== routeLocale) {
      void setLocaleSafe(routeLocale)
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
