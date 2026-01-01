// https://nuxt.com/docs/api/configuration/nuxt-config
const apiBase = process.env.NUXT_API_BASE_URL || 'http://localhost:8080'

export default defineNuxtConfig({
  ssr: true,
  compatibilityDate: '2026-01-01',
  devtools: { enabled: true },
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
  modules: ['@nuxtjs/tailwindcss', '@nuxtjs/color-mode', 'shadcn-nuxt'],
  tailwindcss: {
    cssPath: '~/layers/shared/assets/css/tailwind.css'
  },
  runtimeConfig: {
    public: {
      apiBase: '/api'
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
    prefix: '',
    componentDir: './layers/shared/components/ui'
  }
})
