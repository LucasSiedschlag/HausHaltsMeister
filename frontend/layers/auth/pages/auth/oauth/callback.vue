<script setup lang="ts">

definePageMeta({ layout: false })

const { refresh } = useAuth()
const error = ref('')

onMounted(async () => {
  try {
    await refresh()
    await navigateTo('/')
  } catch {
    error.value = 'Não foi possível concluir o login. Tente novamente.'
  }
})

const goToLogin = () => {
  navigateTo('/auth/login')
}
</script>

<template>
  <div class="bg-muted flex min-h-svh items-center justify-center p-6">
    <div class="w-full max-w-sm rounded-lg border bg-card p-6 text-center text-sm text-muted-foreground">
      <p v-if="!error">Finalizando login...</p>
      <div v-else class="space-y-4">
        <p class="text-destructive">{{ error }}</p>
        <button
          class="mx-auto inline-flex items-center justify-center rounded-md border border-input px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-accent"
          type="button"
          @click="goToLogin"
        >
          Voltar ao login
        </button>
      </div>
    </div>
  </div>
</template>
