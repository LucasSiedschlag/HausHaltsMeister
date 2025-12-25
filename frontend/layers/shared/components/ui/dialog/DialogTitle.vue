<script setup lang="ts">
import type { DialogTitleProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { DialogTitle } from "reka-ui"
import { cn } from "~/layers/shared/utils/cn"

const props = defineProps<DialogTitleProps & { class?: HTMLAttributes["class"] }>()
</script>

<template>
  <DialogTitle :class="cn('text-lg font-semibold leading-none tracking-tight', props.class)">
    <slot />
  </DialogTitle>
</template>
