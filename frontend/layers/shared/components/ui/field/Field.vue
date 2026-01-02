<template>
  <div :class="cn('grid gap-2', attrs.class as string)" v-bind="attrs">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { cn } from '@shared/utils/cn'

const attrs = useAttrs()
</script>
