<script setup lang="ts">
import { Button } from '@shared/components/ui/button'
import { Separator } from '@shared/components/ui/separator'
import { SidebarTrigger } from '@shared/components/ui/sidebar'
import SearchForm from './SearchForm.vue'
</script>

<template>
  <header class="sticky top-0 z-30 flex h-16 items-center gap-4 border-b border-border/60 bg-background/80 px-6 backdrop-blur">
    <SidebarTrigger class="lg:hidden" />
    <Separator orientation="vertical" class="h-6 lg:hidden" />
    <div class="flex-1">
      <SearchForm />
    </div>
    <div class="flex items-center gap-2">
      <Button variant="outline">Exportar</Button>
      <Button>Nova transação</Button>
    </div>
  </header>
</template>
