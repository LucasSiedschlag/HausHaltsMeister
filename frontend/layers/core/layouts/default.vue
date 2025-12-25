<template>
  <div class="min-h-screen w-full bg-background text-foreground">
    <AppSidebar />
    <div class="flex min-h-screen flex-col md:pl-64">
      <AppHeader />
      <main class="flex-1">
        <div class="mx-auto w-full max-w-6xl px-4 py-6 md:px-6 lg:px-8">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppSidebar from '~/layers/core/components/layout/AppSidebar.vue'
import AppHeader from '~/layers/core/components/layout/AppHeader.vue'
</script>
