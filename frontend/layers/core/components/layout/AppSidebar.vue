<template>
  <aside class="fixed inset-y-0 left-0 z-30 hidden w-64 flex-col border-r bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80 md:flex">
    <div class="flex h-14 items-center border-b px-4 md:h-16">
      <NuxtLink to="/" class="flex items-center gap-2 text-base font-semibold">
        <GalleryVerticalEnd class="h-5 w-5" />
        <span>HausHaltsMeister</span>
      </NuxtLink>
    </div>
    <div class="flex-1 overflow-y-auto px-3 py-4">
      <p class="mb-3 text-xs font-semibold uppercase tracking-widest text-muted-foreground">Main</p>
      <nav class="space-y-1 text-sm font-medium">
        <NuxtLink
          v-for="item in navigation"
          :key="item.href"
          :to="item.href"
          class="flex items-center gap-3 rounded-md px-3 py-2 text-muted-foreground transition-colors hover:bg-muted/60 hover:text-foreground"
          active-class="bg-primary/10 text-primary"
        >
          <component :is="getIcon(item.icon)" class="h-4 w-4" />
          {{ item.title }}
        </NuxtLink>
      </nav>
    </div>
    <div class="mt-auto p-4">
      <div class="rounded-lg border bg-card p-4 text-card-foreground shadow-sm">
        <p class="text-sm font-semibold">Upgrade to Pro</p>
        <p class="mt-1 text-sm text-muted-foreground">
          Unlock all features and get unlimited access to our support team.
        </p>
        <Button size="sm" class="mt-4 w-full">
          Upgrade
        </Button>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { mainNavigation } from '~/layers/core/utils/navigation'
import * as LucideIcons from 'lucide-vue-next'
import { GalleryVerticalEnd } from 'lucide-vue-next'

const navigation = mainNavigation

function getIcon(name: string | undefined) {
  if (!name) return 'Circle'
  // @ts-ignore
  return LucideIcons[name] || LucideIcons.Circle
}
</script>
