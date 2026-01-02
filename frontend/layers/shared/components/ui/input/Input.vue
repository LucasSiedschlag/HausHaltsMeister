<template>
  <input
    :type="type"
    :value="modelValue"
    :class="cn('flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm transition-[box-shadow,border-color] duration-200 placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-0 disabled:cursor-not-allowed disabled:opacity-50', attrs.class as string)"
    @input="onInput"
    v-bind="attrs"
  />
</template>

<script setup lang="ts">
import { cn } from '@shared/utils/cn'

const attrs = useAttrs()

defineProps<{
  type?: string
  modelValue?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const onInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}
</script>
