<template>
  <div :class="cn('relative flex items-center justify-center', attrs.class as string)" v-bind="attrs">
    <div class="h-px w-full bg-border" />
    <div
      v-if="$slots.default"
      class="absolute px-2 text-xs uppercase tracking-[0.2em] text-muted-foreground"
      data-slot="field-separator-content"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { cn } from '@shared/utils'

const attrs = useAttrs()
</script>
