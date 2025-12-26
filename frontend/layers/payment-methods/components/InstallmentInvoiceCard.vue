<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '~/layers/shared/components/ui/button'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/layers/shared/components/ui/table'
import MonthPicker from './MonthPicker.vue'
import type { InvoiceEntry, InvoiceSummary } from '../types/installment'
import type { PaymentMethod } from '../types/payment-method'

interface Props {
  paymentMethods: PaymentMethod[]
  paymentMethodsLoading?: boolean
  invoice?: InvoiceSummary | null
  loading?: boolean
  error?: string | null
  monthValue?: string
  cardId?: string
}

const props = withDefaults(defineProps<Props>(), {
  paymentMethods: () => [],
  paymentMethodsLoading: false,
  invoice: null,
  loading: false,
  error: null,
  monthValue: '',
  cardId: '',
})

const emit = defineEmits<{
  'update:monthValue': [value: string]
  'update:cardId': [value: string]
  fetch: []
}>()

const cardOptions = computed(() =>
  props.paymentMethods.filter((method) => method.kind === 'CREDIT_CARD' && method.is_active),
)
const selectedCard = computed(() => props.paymentMethods.find((method) => String(method.id) === props.cardId))

function remainingTotal(entry: InvoiceEntry) {
  const match = entry.title.match(/\((\d+)\/(\d+)\)\s*$/)
  if (!match) return null
  const current = Number(match[1])
  const total = Number(match[2])
  if (!Number.isFinite(current) || !Number.isFinite(total) || total <= 0) return null
  const remainingCount = Math.max(total - current, 0)
  return entry.amount * remainingCount
}

const invoiceRows = computed(() =>
  props.invoice?.entries.map((entry) => ({
    ...entry,
    remaining_total: remainingTotal(entry),
  })) || [],
)

const limitRemaining = computed(() => {
  const limit = selectedCard.value?.credit_limit
  if (!limit || !props.invoice) return null
  return limit - props.invoice.total_remaining
})

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="text-base font-semibold">Fatura do cartão</h2>
          <p class="text-sm text-muted-foreground">Confira o total mensal e as parcelas do período.</p>
        </div>
        <Button variant="outline" :disabled="props.loading" @click="emit('fetch')">Atualizar fatura</Button>
      </div>
    </div>

    <div class="space-y-4 px-6 py-6">
      <div class="grid gap-4 md:grid-cols-2">
        <div class="flex flex-col gap-2">
          <Label for="invoice-card">Cartão</Label>
          <Select
            id="invoice-card"
            :model-value="props.cardId"
            :disabled="props.loading || props.paymentMethodsLoading"
            @update:model-value="emit('update:cardId', $event)"
          >
            <option value="">Selecione um cartão</option>
            <option v-for="method in cardOptions" :key="method.id" :value="String(method.id)">
              {{ method.name }}
            </option>
          </Select>
        </div>

        <div class="flex flex-col gap-2">
          <Label for="invoice-month">Mês de referência</Label>
          <MonthPicker
            id="invoice-month"
            :model-value="props.monthValue"
            :disabled="props.loading"
            @update:model-value="emit('update:monthValue', $event)"
          />
        </div>
      </div>

      <div v-if="props.error" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
        {{ props.error }}
      </div>

      <div v-if="props.invoice" class="grid gap-4 rounded-md border bg-muted/20 p-4 md:grid-cols-3">
        <div>
          <p class="text-sm text-muted-foreground">Total do mês</p>
          <p class="text-2xl font-semibold text-foreground">{{ formatCurrency(props.invoice.total) }}</p>
        </div>
        <div>
          <p class="text-sm text-muted-foreground">Total restante</p>
          <p class="text-2xl font-semibold text-foreground">
            {{ formatCurrency(props.invoice.total_remaining) }}
          </p>
        </div>
        <div>
          <p class="text-sm text-muted-foreground">Limite restante</p>
          <p class="text-2xl font-semibold text-foreground">
            {{ limitRemaining !== null ? formatCurrency(limitRemaining) : '-' }}
          </p>
        </div>
      </div>

      <div v-if="props.loading" class="space-y-3">
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
      </div>

      <div v-else-if="props.invoice && !props.invoice.entries.length" class="text-sm text-muted-foreground">
        Nenhuma parcela encontrada para este mês.
      </div>

      <div v-else-if="props.invoice" class="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Data</TableHead>
              <TableHead>Descrição</TableHead>
              <TableHead>Categoria</TableHead>
              <TableHead>Restante</TableHead>
              <TableHead class="text-right">Valor</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="entry in invoiceRows" :key="`${entry.title}-${entry.date}`">
              <TableCell class="text-sm text-muted-foreground">{{ entry.date }}</TableCell>
              <TableCell class="font-medium">{{ entry.title }}</TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ entry.category_name }}</TableCell>
              <TableCell class="text-sm text-muted-foreground">
                {{ entry.remaining_total !== null ? formatCurrency(entry.remaining_total) : '-' }}
              </TableCell>
              <TableCell class="text-right">{{ formatCurrency(entry.amount) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </div>
  </div>
</template>
