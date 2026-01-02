<template>
  <h3 :class="cn('text-lg font-semibold leading-none tracking-tight', attrs.class as string)" v-bind="attrs">
    <slot />
  </h3>
</template>

<script setup lang="ts">
import { cn } from '@shared/utils/cn'

const attrs = useAttrs()
</script>
