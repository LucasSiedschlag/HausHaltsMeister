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
    './layers/investments',
    './layers/creditcard',
    './layers/reports'
  ],
  modules: ['@nuxtjs/tailwindcss', '@nuxtjs/color-mode', 'shadcn-nuxt', 'notivue/nuxt'],
  css: [
    join(rootDir, 'layers/shared/assets/css/tailwind.css'),
    'notivue/notification.css',
    'notivue/notification-progress.css',
    'notivue/animations.css'
  ],
  runtimeConfig: {
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
  shadcn: {
    prefix: 'Ui',
    componentDir: './layers/shared/components/ui'
  }
})
