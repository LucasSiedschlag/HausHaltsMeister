<script setup lang="ts">
import { darkTheme, lightTheme, NotificationProgress, updateConfig } from 'notivue'
import { usePreferences } from '@shared/composables/usePreferences'

const colorMode = useColorMode()
const { preferences } = usePreferences()

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
</template>
