<script setup lang="ts">
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@shared/components/ui/card'

defineProps<{
  title: string
  description?: string
  loading?: boolean
  error?: string | null
  empty?: boolean
  loadingMessage?: string
  emptyMessage?: string
}>()
</script>

<template>
  <Card>
    <CardHeader class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <CardTitle>{{ title }}</CardTitle>
        <CardDescription v-if="description">{{ description }}</CardDescription>
      </div>
      <div v-if="$slots.toolbar" class="w-full md:w-auto">
        <slot name="toolbar" />
      </div>
    </CardHeader>
    <CardContent>
      <div v-if="loading" class="py-6 text-sm text-muted-foreground">
        {{ loadingMessage }}
      </div>
      <div v-else-if="error" class="py-6 text-sm text-destructive">
        {{ error }}
      </div>
      <div v-else-if="empty" class="py-6 text-sm text-muted-foreground">
        <slot name="empty">
          {{ emptyMessage }}
        </slot>
      </div>
      <div v-else>
        <slot />
      </div>
    </CardContent>
  </Card>
</template>
