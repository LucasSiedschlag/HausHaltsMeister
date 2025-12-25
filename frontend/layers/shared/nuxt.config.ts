import { createResolver } from '@nuxt/kit'

const { resolve } = createResolver(import.meta.url)

export default defineNuxtConfig({
  components: [
    {
      path: resolve('components'),
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
    componentDir: resolve('./components/ui')
  },
  css: [
    resolve('./assets/css/tailwind.css')
  ],
  tailwindcss: {
    viewer: false,
  }
})

