<template>
  <p :class="cn('text-sm text-muted-foreground', attrs.class as string)" v-bind="attrs">
    <slot />
  </p>
</template>

<script setup lang="ts">
import { cn } from '@shared/utils'

const attrs = useAttrs()
</script>
