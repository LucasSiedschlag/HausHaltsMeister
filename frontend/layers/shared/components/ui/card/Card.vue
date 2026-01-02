<template>
  <div :class="cn('rounded-lg border bg-card text-card-foreground shadow-sm', attrs.class as string)" v-bind="attrs">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { cn } from '@shared/utils/cn'

const attrs = useAttrs()
</script>
