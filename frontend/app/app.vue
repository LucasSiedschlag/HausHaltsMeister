<script setup lang="ts">
import { darkTheme, lightTheme, NotificationProgress, updateConfig } from 'notivue'
import { usePreferences } from '@shared/composables/usePreferences'
import { Button } from '@shared/components/ui/button'
import { useSessionBlock } from '@shared/composables/useSessionBlock'

const colorMode = useColorMode()
const { preferences } = usePreferences()
const { sessionExpired, acknowledgeSessionExpired } = useAuth()
const { blocked } = useSessionBlock()

const sessionBlocked = computed(() => blocked.value && sessionExpired.value)

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
  <div v-if="sessionBlocked" class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 px-6 backdrop-blur">
    <div class="w-full max-w-md rounded-lg border border-border/60 bg-background p-6 shadow-lg">
      <h2 class="text-lg font-semibold">Sessão expirada</h2>
      <p class="mt-2 text-sm text-muted-foreground">
        Seu acesso expirou. Faça login novamente para continuar.
      </p>
      <div class="mt-6 flex justify-end">
        <Button @click="acknowledgeSessionExpired">Ir para o login</Button>
      </div>
    </div>
  </div>
</template>
