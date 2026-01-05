// https://nuxt.com/docs/api/configuration/nuxt-config
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

const apiBase = process.env.NUXT_API_BASE_URL || 'http://localhost:8080'
const rootDir = dirname(fileURLToPath(import.meta.url))

export default defineNuxtConfig({
  ssr: true,
  compatibilityDate: '2026-01-01',
  devtools: { enabled: true },
  alias: {
    '@shared': join(rootDir, 'layers/shared')
  },
  extends: [
    './layers/shared',
    './layers/core',
    './layers/auth',
    './layers/ledgers',
    './layers/accounts',
    './layers/categories',
    './layers/journal',
    './layers/budget',
    './layers/core',
    './layers/investments',
    './layers/creditcard',
    './layers/reports'
  ],
  modules: ['@nuxtjs/tailwindcss', '@nuxtjs/color-mode', '@nuxtjs/i18n', 'shadcn-nuxt', 'notivue/nuxt'],
  css: [
    join(rootDir, 'layers/shared/assets/css/tailwind.css'),
    'notivue/notification.css',
    'notivue/notification-progress.css',
    'notivue/animations.css'
  ],
  runtimeConfig: {
    apiBase,
    public: {
      apiBase: '/api',
      oauthBase: process.env.NUXT_PUBLIC_OAUTH_BASE_URL || '',
      refreshCookieName: process.env.NUXT_PUBLIC_REFRESH_COOKIE_NAME || 'hhm_refresh'
    }
  },
  nitro: {
    routeRules: {
      '/api/**': {
        proxy: `${apiBase}/**`
      },
      // Disable SSR for auth forms - eliminates hydration issues
      // Keep /auth/oauth/callback with SSR to properly read cookies
      '/auth/login': {
        ssr: false
      },
      '/auth/signup': {
        ssr: false
      },
      '/auth/forgot-password': {
        ssr: false
      }
    }
  },
  colorMode: {
    classSuffix: '',
    preference: 'system',
    fallback: 'light',
    storage: 'cookie',
    storageKey: 'hhm-color-mode'
  },
  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'pt-BR',
    detectBrowserLanguage: {
      alwaysRedirect: false,
      redirectOn: 'root',
      useCookie: true,
      cookieKey: 'hhm_locale',
      fallbackLocale: 'pt-BR'
    },
    restructureDir: '.',
    compilation: {
      strictMessage: false
    },
    langDir: 'layers/shared/i18n',
    locales: [
      { code: 'pt-BR', iso: 'pt-BR', name: 'Português (Brasil)', file: 'pt-BR.json' },
      { code: 'en-US', iso: 'en-US', name: 'English (US)', file: 'en-US.json' }
    ]
  },
  shadcn: {
    prefix: 'Ui',
    componentDir: './layers/shared/components/ui'
  }
})
