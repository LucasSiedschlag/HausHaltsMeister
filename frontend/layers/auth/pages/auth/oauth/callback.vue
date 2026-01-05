<script setup lang="ts">

definePageMeta({ layout: false, ssr: false })

const { accessToken, user, refresh } = useAuth()
const error = ref('')
const { t } = useI18n()

onMounted(async () => {
  try {
    // OAuth flow sets the refresh cookie on the backend
    // We just need to call refresh() to get a new access token
    if (!accessToken.value) {
      await refresh()
    }

    // Verify we have a valid session
    if (accessToken.value && user.value) {
      await navigateTo('/', { replace: true })
    } else {
      error.value = t('auth.oauth.error')
    }
  } catch {
    error.value = t('auth.oauth.error')
  }
})

const goToLogin = () => {
  navigateTo('/auth/login')
}
</script>

<template>
  <div class="bg-muted flex min-h-svh items-center justify-center p-6">
    <div class="w-full max-w-sm rounded-lg border bg-card p-6 text-center text-sm text-muted-foreground">
      <p v-if="!error">{{ t('auth.oauth.finishing') }}</p>
      <div v-else class="space-y-4">
        <p class="text-destructive">{{ error }}</p>
        <button
          class="mx-auto inline-flex items-center justify-center rounded-md border border-input px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-accent"
          type="button"
          @click="goToLogin"
        >
          {{ t('auth.oauth.backToLogin') }}
        </button>
      </div>
    </div>
  </div>
</template>
