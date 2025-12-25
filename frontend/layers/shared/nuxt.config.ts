// frontend/layers/shared/nuxt.config.ts
export default defineNuxtConfig({
  components: [
    {
      path: 'components',
      extensions: ['.vue'],
      pathPrefix: false 
    }
  ],
  modules: [
    '@nuxtjs/tailwindcss',
    'shadcn-nuxt'
  ],
  shadcn: {
    prefix: '',
    componentDir: './components/ui'
  },
  tailwindcss: {
    cssPath: './assets/css/tailwind.css',
    configPath: 'tailwind.config.ts'
  }
})
