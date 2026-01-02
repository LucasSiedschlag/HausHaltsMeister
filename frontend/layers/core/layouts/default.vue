<script setup lang="ts">
import { SidebarInset, SidebarProvider } from '@shared/components/ui/sidebar'
import AppSidebar from '../components/AppSidebar.vue'
import SiteHeader from '../components/SiteHeader.vue'
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset>
      <SiteHeader />
      <main class="flex-1 px-6 py-6">
        <div class="mx-auto w-full max-w-6xl">
          <slot />
        </div>
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
