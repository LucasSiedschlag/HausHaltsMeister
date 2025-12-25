<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "~/layers/shared/utils/cn"

interface Props {
  modelValue?: string
  class?: HTMLAttributes["class"]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  "update:modelValue": [value: string]
}>()
</script>

<template>
  <select
    v-bind="$attrs"
    :value="props.modelValue"
    :class="cn(
      'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      props.class,
    )"
    @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
  >
    <slot />
  </select>
</template>
