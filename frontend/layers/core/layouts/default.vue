<template>
  <div class="min-h-screen bg-background text-foreground">
    <AppHeader />
    <main class="mx-auto w-full max-w-6xl px-6 py-6">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import AppHeader from '~/components/layout/AppHeader.vue'
</script>
