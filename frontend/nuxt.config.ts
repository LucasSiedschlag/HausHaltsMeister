// frontend/nuxt.config.ts
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true },
  css: [
    '~/layers/shared/assets/css/tailwind.css'
  ],
  extends: [
    './layers/shared',
    './layers/core',
    './layers/categories',
    './layers/budget',
    './layers/dashboard'
  ],
  modules: [
    '@nuxtjs/tailwindcss',
    'shadcn-nuxt',
    '@nuxtjs/color-mode'
  ],
  colorMode: {
    classSuffix: '',
    preference: 'system',
    fallback: 'light',
    disableTransition: true
  },
  runtimeConfig: {
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || '/api',
      apiToken: process.env.NUXT_PUBLIC_API_TOKEN || '',
      apiTimeoutMs: Number(process.env.NUXT_PUBLIC_API_TIMEOUT_MS || 15000),
    }
  },
  nitro: {
    devProxy: {
      '/api': {
        target: process.env.NUXT_PUBLIC_API_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
})
