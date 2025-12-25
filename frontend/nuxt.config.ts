// frontend/nuxt.config.ts
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true },
  extends: [
    './layers/shared',
    './layers/core',
    './layers/dashboard'
  ],
  modules: [
    '@nuxtjs/tailwindcss',
    'shadcn-nuxt'
  ], 
  // Global config if needed
  shadcn: {
    // Explicitly pointing to shared layer for correct resolution if needed,
    // though the layer config should handle it.
    componentDir: './layers/shared/components/ui' 
  }
})
