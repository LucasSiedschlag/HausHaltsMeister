<template>
  <label
    :for="forId"
    :class="cn('text-sm font-medium leading-none text-foreground', attrs.class as string)"
    v-bind="attrs"
  >
    <slot />
  </label>
</template>

<script setup lang="ts">
import { cn } from '@shared/utils'

const attrs = useAttrs()

defineProps<{
  forId?: string
}>()
</script>
