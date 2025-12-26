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
import type { PicuinhaCase } from '../types/picuinha'

interface CaseRow extends PicuinhaCase {
  person_name?: string
}

interface Props {
  cases: CaseRow[]
  loading: boolean
  error?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  error: null,
})

const emit = defineEmits<{
  create: []
  viewInstallments: [picCase: CaseRow]
  remove: [picCase: CaseRow]
  retry: []
}>()

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

function statusLabel(status: PicuinhaCase['status']) {
  if (status === 'PAID') return 'Paga'
  if (status === 'RECURRING') return 'Recorrente'
  return 'Aberta'
}

function statusClass(status: PicuinhaCase['status']) {
  if (status === 'PAID') {
    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
  }
  if (status === 'RECURRING') {
    return 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:border-sky-500/40 dark:bg-sky-500/20 dark:text-sky-300'
  }
  return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:border-amber-500/40 dark:bg-amber-500/20 dark:text-amber-300'
}

function typeLabel(value: PicuinhaCase['case_type']) {
  switch (value) {
    case 'ONE_OFF':
      return 'Única'
    case 'INSTALLMENT':
      return 'Parcelada'
    case 'CARD_INSTALLMENT':
      return 'Cartão'
    case 'RECURRING':
      return 'Recorrente'
    default:
      return value
  }
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold">Picuinhas</h2>
          <p class="text-sm text-muted-foreground">Lançamentos individuais por pessoa.</p>
        </div>
        <Button @click="emit('create')">Nova picuinha</Button>
      </div>
    </div>

    <div v-if="props.loading" class="space-y-4 px-6 py-6">
      <div class="h-8 w-40 animate-pulse rounded-md bg-muted" />
      <div class="space-y-3">
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

    <div v-else-if="!props.cases.length" class="px-6 py-10 text-center">
      <p class="text-sm text-muted-foreground">Sem picuinhas cadastradas.</p>
      <Button class="mt-4" @click="emit('create')">Criar primeira picuinha</Button>
    </div>

    <div v-else class="px-2 py-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Pessoa</TableHead>
            <TableHead>Título</TableHead>
            <TableHead>Tipo</TableHead>
            <TableHead>Parcelas</TableHead>
            <TableHead>Restante</TableHead>
            <TableHead>Status</TableHead>
            <TableHead class="text-right">Ações</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="picCase in props.cases" :key="picCase.id">
            <TableCell class="text-sm text-muted-foreground">{{ picCase.person_name || '-' }}</TableCell>
            <TableCell class="font-medium">{{ picCase.title }}</TableCell>
            <TableCell class="text-sm text-muted-foreground">{{ typeLabel(picCase.case_type) }}</TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ picCase.installments_paid }}/{{ picCase.installments_total }}
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ formatCurrency(picCase.amount_remaining) }}
            </TableCell>
            <TableCell>
              <Badge variant="outline" :class="statusClass(picCase.status)">
                {{ statusLabel(picCase.status) }}
              </Badge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-2">
                <Button variant="secondary" size="sm" @click="emit('viewInstallments', picCase)">
                  Parcelas
                </Button>
                <Button variant="destructive" size="sm" @click="emit('remove', picCase)">
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
