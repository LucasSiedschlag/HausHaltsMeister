<template>
  <div :class="cn('relative flex h-px w-full items-center justify-center', attrs.class as string)" v-bind="attrs">
    <div class="h-px w-full bg-border" />
    <div v-if="$slots.default" class="absolute bg-background px-2 text-xs text-muted-foreground">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { cn } from '@shared/utils/cn'

const attrs = useAttrs()
</script>
