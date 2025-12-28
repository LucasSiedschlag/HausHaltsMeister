<script setup lang="ts">
import { Badge } from '~/layers/shared/components/ui/badge'
import { Button } from '~/layers/shared/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/layers/shared/components/ui/table'
import type { CashflowEntry } from '../types/cashflow'

interface Props {
  entries: CashflowEntry[]
  loading: boolean
  error?: string | null
  title: string
  description?: string
  emptyMessage?: string
  emptyActionLabel?: string
}

const props = withDefaults(defineProps<Props>(), {
  error: null,
  description: '',
  emptyMessage: 'Sem lançamentos para o mês selecionado.',
  emptyActionLabel: 'Criar primeiro lançamento',
})

const emit = defineEmits<{
  create: []
  edit: [entry: CashflowEntry]
  remove: [entry: CashflowEntry]
  reverse: [entry: CashflowEntry]
  retry: []
}>()

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

function parseLocalDate(value: string) {
  const [year, month, day] = value.split('-').map(Number)
  if (year && month && day) {
    return new Date(year, month - 1, day)
  }
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return null
  return parsed
}

function formatDate(value: string) {
  const parsed = parseLocalDate(value)
  if (!parsed) return value
  return parsed.toLocaleDateString('pt-BR')
}

function directionLabel(direction: string) {
  return direction === 'IN' ? 'Entrada' : 'Saída'
}

function directionBadgeClass(direction: string) {
  if (direction === 'IN') {
    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
  }
  return 'border-red-500/30 bg-red-500/10 text-red-700 dark:border-red-500/40 dark:bg-red-500/20 dark:text-red-300'
}

function fixedBadgeClass() {
  return 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:border-sky-500/40 dark:bg-sky-500/20 dark:text-sky-300'
}

function reversalBadgeClass() {
  return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:border-amber-500/40 dark:bg-amber-500/20 dark:text-amber-300'
}

function amountClass(entry: CashflowEntry) {
  if (entry.direction === 'IN') {
    return 'text-emerald-600 dark:text-emerald-400'
  }
  return 'text-red-600 dark:text-red-400'
}

function isInstallmentEntry(entry: CashflowEntry) {
  return entry.id < 0
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold">{{ props.title }}</h2>
          <p v-if="props.description" class="text-sm text-muted-foreground">{{ props.description }}</p>
        </div>
        <Button @click="emit('create')">Novo lançamento</Button>
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

    <div v-else-if="!props.entries.length" class="px-6 py-10 text-center">
      <p class="text-sm text-muted-foreground">{{ props.emptyMessage }}</p>
      <Button class="mt-4" @click="emit('create')">{{ props.emptyActionLabel }}</Button>
    </div>

    <div v-else class="px-2 py-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Data</TableHead>
            <TableHead>Título</TableHead>
            <TableHead>Categoria</TableHead>
            <TableHead>Meio</TableHead>
            <TableHead>Tipo</TableHead>
            <TableHead>Valor</TableHead>
            <TableHead class="text-right">Ações</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="entry in props.entries" :key="entry.id">
            <TableCell class="text-sm text-muted-foreground">{{ formatDate(entry.date) }}</TableCell>
            <TableCell class="font-medium">{{ entry.title }}</TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ entry.category_name || `Categoria #${entry.category_id}` }}
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ entry.payment_method_name || `Meio #${entry.payment_method_id}` }}
            </TableCell>
            <TableCell>
              <div class="flex flex-wrap items-center gap-2">
                <Badge variant="outline" :class="directionBadgeClass(entry.direction)">
                  {{ directionLabel(entry.direction) }}
                </Badge>
                <Badge v-if="entry.is_fixed" variant="outline" :class="fixedBadgeClass()">
                  Fixo
                </Badge>
                <Badge v-if="entry.reversal_of_entry_id" variant="outline" :class="reversalBadgeClass()">
                  Estorno
                </Badge>
                <Badge v-if="isInstallmentEntry(entry)" variant="outline" class="border-indigo-500/30 bg-indigo-500/10 text-indigo-700 dark:border-indigo-500/40 dark:bg-indigo-500/20 dark:text-indigo-300">
                  Parcela
                </Badge>
              </div>
            </TableCell>
            <TableCell class="text-sm font-semibold" :class="amountClass(entry)">
              {{ formatCurrency(entry.amount) }}
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-2">
                <Button
                  variant="secondary"
                  size="sm"
                  :disabled="isInstallmentEntry(entry)"
                  :title="isInstallmentEntry(entry) ? 'Parcela gerada automaticamente.' : undefined"
                  @click="emit('edit', entry)"
                >
                  Editar
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  :disabled="Boolean(entry.reversal_of_entry_id) || isInstallmentEntry(entry)"
                  :title="entry.reversal_of_entry_id ? 'Lançamento já estornado.' : (isInstallmentEntry(entry) ? 'Parcela gerada automaticamente.' : undefined)"
                  @click="emit('reverse', entry)"
                >
                  Estornar
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  class="text-destructive hover:text-destructive"
                  :disabled="isInstallmentEntry(entry)"
                  :title="isInstallmentEntry(entry) ? 'Parcela gerada automaticamente.' : undefined"
                  @click="emit('remove', entry)"
                >
                  Excluir
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
