<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Switch } from '~/layers/shared/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/layers/shared/components/ui/table'
import type { PicuinhaCase, PicuinhaCaseInstallment } from '../types/picuinha'

interface Props {
  open: boolean
  picCase?: PicuinhaCase | null
  installments: PicuinhaCaseInstallment[]
  loading?: boolean
  error?: string | null
  updatingId?: number | null
}

const props = withDefaults(defineProps<Props>(), {
  picCase: null,
  loading: false,
  error: null,
  updatingId: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  updateInstallment: [payload: { id: number; is_paid: boolean; extra_amount: number }]
}>()

const extras = ref<Record<number, string>>({})

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

const title = computed(() => (props.picCase ? `Parcelas · ${props.picCase.title}` : 'Parcelas'))

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

function totalAmount(installment: PicuinhaCaseInstallment) {
  const extra = Number(extras.value[installment.id] ?? installment.extra_amount ?? 0)
  return installment.amount + (Number.isFinite(extra) ? extra : 0)
}

function syncExtras() {
  const next: Record<number, string> = {}
  props.installments.forEach((installment) => {
    next[installment.id] = String(installment.extra_amount ?? 0)
  })
  extras.value = next
}

function updateExtra(installment: PicuinhaCaseInstallment) {
  const value = Number(extras.value[installment.id] || 0)
  if (!Number.isFinite(value)) return
  emit('updateInstallment', {
    id: installment.id,
    is_paid: installment.is_paid,
    extra_amount: value,
  })
}

function togglePaid(installment: PicuinhaCaseInstallment, nextValue: boolean) {
  const value = Number(extras.value[installment.id] || 0)
  emit('updateInstallment', {
    id: installment.id,
    is_paid: nextValue,
    extra_amount: Number.isFinite(value) ? value : 0,
  })
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      syncExtras()
    }
  },
)

watch(
  () => props.installments,
  () => {
    if (props.open) {
      syncExtras()
    }
  },
)
</script>

<template>
  <Sheet :open="props.open" @update:open="emit('update:open', $event)">
    <SheetContent side="right" class="w-full sm:max-w-2xl">
      <SheetHeader>
        <SheetTitle>{{ title }}</SheetTitle>
      </SheetHeader>

      <div class="mt-6 max-h-[70vh] space-y-4 overflow-y-auto pr-2">
        <div v-if="props.error" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
          {{ props.error }}
        </div>

        <div v-if="props.loading" class="space-y-3">
          <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
          <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
          <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
        </div>

        <div v-else-if="!props.installments.length" class="text-sm text-muted-foreground">
          Nenhuma parcela encontrada.
        </div>

        <div v-else class="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Parcela</TableHead>
                <TableHead>Vencimento</TableHead>
                <TableHead>Valor</TableHead>
                <TableHead>Juros</TableHead>
                <TableHead>Total</TableHead>
                <TableHead class="text-right">Pago</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="installment in props.installments" :key="installment.id">
                <TableCell class="font-medium">{{ installment.installment_number }}</TableCell>
                <TableCell class="text-sm text-muted-foreground">{{ installment.due_date }}</TableCell>
                <TableCell class="text-sm text-muted-foreground">{{ formatCurrency(installment.amount) }}</TableCell>
                <TableCell>
                  <Input
                    v-model="extras[installment.id]"
                    type="number"
                    min="0"
                    step="0.01"
                    class="h-8 w-24"
                    :disabled="props.updatingId === installment.id"
                    @blur="updateExtra(installment)"
                  />
                </TableCell>
                <TableCell class="text-sm text-muted-foreground">{{ formatCurrency(totalAmount(installment)) }}</TableCell>
                <TableCell class="text-right">
                  <Switch
                    :model-value="installment.is_paid"
                    :disabled="props.updatingId === installment.id"
                    @update:model-value="togglePaid(installment, $event)"
                  />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </div>

      <div class="mt-6 flex justify-end">
        <Button variant="outline" @click="emit('update:open', false)">Fechar</Button>
      </div>
    </SheetContent>
  </Sheet>
</template>
