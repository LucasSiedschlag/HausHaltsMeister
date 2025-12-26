<script setup lang="ts">
import { Badge } from '~/layers/shared/components/ui/badge'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/layers/shared/components/ui/table'

interface BudgetRow {
  category_id: number
  category_name: string
  planned_amount: number
  actual_amount: number
  target_percent: number
  inactive?: boolean
}

interface Props {
  rows: BudgetRow[]
  loading: boolean
  error?: string | null
  saveState?: 'ready' | 'adjust'
  saving?: boolean
  saveLabel?: string
}

const props = withDefaults(defineProps<Props>(), {
  error: null,
  saveState: 'adjust',
  saving: false,
  saveLabel: 'Salvar orçamento',
})

const emit = defineEmits<{
  updatePercent: [payload: { categoryId: number; value: number }]
  save: []
  retry: []
}>()

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

function deltaBadgeClass(planned: number, actual: number) {
  const delta = planned - actual
  if (delta >= 0) {
    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
  }
  return 'border-red-500/30 bg-red-500/10 text-red-700 dark:border-red-500/40 dark:bg-red-500/20 dark:text-red-300'
}

function handlePercentInput(categoryId: number, value: string) {
  const parsed = Number(value)
  const normalized = Number.isNaN(parsed) ? 0 : Math.max(0, Math.min(100, Math.round(parsed)))
  emit('updatePercent', { categoryId, value: normalized })
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex items-center justify-between gap-4">
        <div>
          <h2 class="text-base font-semibold">Distribuição por categoria</h2>
          <p class="text-sm text-muted-foreground">Defina os percentuais para o mês selecionado.</p>
        </div>
        <Button
          variant="outline"
          :disabled="props.saving"
          :class="props.saveState === 'adjust'
            ? 'border-amber-500/40 bg-amber-500/10 text-amber-700 hover:bg-amber-500/20 dark:border-amber-400/50 dark:bg-amber-400/10 dark:text-amber-300'
            : 'border-primary/30 bg-primary/10 text-primary hover:bg-primary/20'"
          @click="emit('save')"
        >
          {{ props.saving ? 'Salvando...' : props.saveLabel }}
        </Button>
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

    <div v-else-if="!props.rows.length" class="px-6 py-10 text-center">
      <p class="text-sm text-muted-foreground">Sem categorias elegíveis para orçamento.</p>
    </div>

    <div v-else class="px-2 py-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Categoria</TableHead>
            <TableHead class="w-32">Percentual</TableHead>
            <TableHead>Planejado</TableHead>
            <TableHead>Realizado</TableHead>
            <TableHead>Saldo</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="row in props.rows" :key="row.category_id" :class="row.inactive ? 'opacity-70' : ''">
            <TableCell class="font-medium">
              <div class="flex flex-wrap items-center gap-2">
                <span>{{ row.category_name }}</span>
                <Badge
                  v-if="row.inactive"
                  variant="outline"
                  class="border-amber-500/40 bg-amber-500/10 text-amber-700 dark:border-amber-400/50 dark:bg-amber-400/10 dark:text-amber-300"
                >
                  Inativa neste mês
                </Badge>
              </div>
            </TableCell>
            <TableCell>
              <div class="flex items-center gap-2">
                <Input
                  :model-value="row.target_percent"
                  type="number"
                  min="0"
                  max="100"
                  step="1"
                  class="h-9 w-20"
                  :disabled="row.inactive"
                  @update:model-value="handlePercentInput(row.category_id, $event)"
                />
                <span class="text-sm text-muted-foreground">%</span>
              </div>
            </TableCell>
            <TableCell>{{ formatCurrency(row.planned_amount) }}</TableCell>
            <TableCell>{{ formatCurrency(row.actual_amount) }}</TableCell>
            <TableCell>
              <Badge variant="outline" :class="deltaBadgeClass(row.planned_amount, row.actual_amount)">
                {{ formatCurrency(row.planned_amount - row.actual_amount) }}
              </Badge>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
