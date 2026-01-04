<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@shared/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <div
    role="group"
    :class="cn('inline-flex -space-x-px overflow-hidden rounded-md border border-border/60 bg-background shadow-sm', props.class)"
  >
    <slot />
  </div>
</template>
