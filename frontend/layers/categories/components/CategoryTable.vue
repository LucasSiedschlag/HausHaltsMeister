<script setup lang="ts">
import { Button } from '~/layers/shared/components/ui/button'
import { Badge } from '~/layers/shared/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/layers/shared/components/ui/table'
import type { Category } from '../types/category'

interface Props {
  categories: Category[]
  loading: boolean
  error?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  error: null,
})

const emit = defineEmits<{
  create: []
  edit: [category: Category]
  remove: [category: Category]
  retry: []
}>()

function directionLabel(direction: string) {
  return direction === 'IN' ? 'Entrada' : 'Saída'
}

function directionBadgeClass(direction: string) {
  if (direction === 'IN') {
    return 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:border-sky-500/40 dark:bg-sky-500/20 dark:text-sky-300'
  }
  return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:border-amber-500/40 dark:bg-amber-500/20 dark:text-amber-300'
}

function statusBadgeClass(isActive: boolean) {
  if (isActive) {
    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
  }
  return 'border-red-500/30 bg-red-500/10 text-red-700 dark:border-red-500/40 dark:bg-red-500/20 dark:text-red-300'
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold">Categorias</h2>
          <p class="text-sm text-muted-foreground">Gerencie entradas e saidas do fluxo.</p>
        </div>
        <Button @click="emit('create')">Nova categoria</Button>
      </div>
    </div>

    <div v-if="props.loading" class="space-y-4 px-6 py-6">
      <div class="h-8 w-40 animate-pulse rounded-md bg-muted" />
      <div class="space-y-3">
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
      </div>
    </div>

    <div v-else-if="props.error" class="px-6 py-6">
      <div class="rounded-md border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
        {{ props.error }}
      </div>
      <Button variant="outline" class="mt-4" @click="emit('retry')">
        Tentar novamente
      </Button>
    </div>

    <div v-else-if="!props.categories.length" class="px-6 py-10 text-center">
      <p class="text-sm text-muted-foreground">Sem categorias cadastradas.</p>
      <Button class="mt-4" @click="emit('create')">Criar primeira categoria</Button>
    </div>

    <div v-else class="px-2 py-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nome</TableHead>
            <TableHead>Direção</TableHead>
            <TableHead>Orcamento</TableHead>
            <TableHead>Status</TableHead>
            <TableHead class="text-right">Ações</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="category in props.categories" :key="category.id">
            <TableCell class="font-medium">{{ category.name }}</TableCell>
            <TableCell>
              <Badge variant="outline" :class="directionBadgeClass(category.direction)">
                {{ directionLabel(category.direction) }}
              </Badge>
            </TableCell>
            <TableCell>
              <span class="text-sm text-muted-foreground">
                {{ category.is_budget_relevant ? 'Sim' : 'Não' }}
              </span>
            </TableCell>
            <TableCell>
              <Badge variant="outline" :class="statusBadgeClass(category.is_active)">
                {{ category.is_active ? 'Ativa' : 'Inativa' }}
              </Badge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-2">
                <Button variant="secondary" size="sm" @click="emit('edit', category)">Editar</Button>
                <Button variant="destructive" size="sm" @click="emit('remove', category)">
                  Desativar
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
