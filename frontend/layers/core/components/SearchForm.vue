<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import { SidebarInput } from '@shared/components/ui/sidebar'
</script>

<template>
  <form class="px-2 pb-2">
    <div class="relative">
      <Search class="pointer-events-none absolute left-2 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <SidebarInput placeholder="Buscar" class="pl-8" />
    </div>
  </form>
</template>
