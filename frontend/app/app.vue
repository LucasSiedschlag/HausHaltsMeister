<script setup lang="ts">
import { darkTheme, lightTheme, NotificationProgress, updateConfig } from 'notivue'
import { usePreferences } from '@shared/composables/usePreferences'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'

const colorMode = useColorMode()
const { preferences } = usePreferences()
const { sessionExpired, acknowledgeSessionExpired } = useAuth()

const notivueTheme = computed(() => {
  const mode = preferences.value?.theme_mode ?? colorMode.preference ?? 'system'
  const resolved = mode === 'system' ? colorMode.value : mode
  return resolved === 'dark' ? darkTheme : lightTheme
})

onMounted(() => {
  updateConfig({
    pauseOnHover: true,
    pauseOnTabChange: false,
    limit: 5,
    position: 'bottom-center',
    notifications: {
      global: {
        duration: 3000
      }
    }
  })
})
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <Notivue v-slot="item">
    <Notification :item="item" :theme="notivueTheme">
      <NotificationProgress :item="item" />
    </Notification>
  </Notivue>
  <ConfirmDialog
    :open="sessionExpired"
    title="Sessão expirada"
    description="Seu acesso expirou. Faça login novamente para continuar."
    confirm-label="Ir para o login"
    cancel-label="Fechar"
    @confirm="acknowledgeSessionExpired"
    @cancel="acknowledgeSessionExpired"
    @update:open="acknowledgeSessionExpired"
  />
</template>
