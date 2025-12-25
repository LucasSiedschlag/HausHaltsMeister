<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "~/layers/shared/utils/cn"

interface Props {
  modelValue?: string | number
  type?: string
  class?: HTMLAttributes["class"]
}

const props = withDefaults(defineProps<Props>(), {
  type: "text",
})

const emit = defineEmits<{
  "update:modelValue": [value: string]
}>()
</script>

<template>
  <input
    v-bind="$attrs"
    :type="props.type"
    :value="modelValue"
    :class="cn(
      'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      props.class,
    )"
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
