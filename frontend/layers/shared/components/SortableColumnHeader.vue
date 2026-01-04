<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { computed } from "vue"
import { ArrowUpDown } from "lucide-vue-next"
import { Button } from "@shared/components/ui/button"
import { cn } from "@shared/utils"

const props = defineProps<{
  label: string
  align?: "left" | "right"
  class?: HTMLAttributes["class"]
}>()

const emit = defineEmits<{
  toggle: []
}>()

const alignClass = computed(() =>
  props.align === "right" ? "justify-end" : "justify-start",
)
</script>

<template>
  <Button
    variant="ghost"
    :class="cn('w-full px-0 pl-0 pr-0 has-[>svg]:px-0', alignClass, props.class)"
    @click="emit('toggle')"
  >
    {{ label }}
    <ArrowUpDown class="ml-2 h-4 w-4" />
  </Button>
</template>
