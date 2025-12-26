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
import type { PaymentMethod, PaymentMethodKind } from '../types/payment-method'

interface Props {
  methods: PaymentMethod[]
  loading: boolean
  error?: string | null
  actionsDisabled?: boolean
  actionsDisabledReason?: string
}

const props = withDefaults(defineProps<Props>(), {
  error: null,
  actionsDisabled: false,
  actionsDisabledReason: 'Funcionalidade pendente no backend.',
})

const emit = defineEmits<{
  create: []
  edit: [method: PaymentMethod]
  remove: [method: PaymentMethod]
  retry: []
}>()

function kindLabel(kind: PaymentMethodKind) {
  switch (kind) {
    case 'CREDIT_CARD':
      return 'Cartão de crédito'
    case 'DEBIT_CARD':
      return 'Cartão de débito'
    case 'CASH':
      return 'Dinheiro'
    case 'PIX':
      return 'PIX'
    case 'BANK_SLIP':
      return 'Boleto'
    default:
      return kind
  }
}

function kindBadgeClass(kind: PaymentMethodKind) {
  switch (kind) {
    case 'CREDIT_CARD':
      return 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:border-sky-500/40 dark:bg-sky-500/20 dark:text-sky-300'
    case 'DEBIT_CARD':
      return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
    case 'PIX':
      return 'border-indigo-500/30 bg-indigo-500/10 text-indigo-700 dark:border-indigo-500/40 dark:bg-indigo-500/20 dark:text-indigo-300'
    case 'CASH':
      return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:border-amber-500/40 dark:bg-amber-500/20 dark:text-amber-300'
    case 'BANK_SLIP':
      return 'border-slate-500/30 bg-slate-500/10 text-slate-700 dark:border-slate-500/40 dark:bg-slate-500/20 dark:text-slate-300'
    default:
      return 'border-muted/50 bg-muted/40 text-muted-foreground'
  }
}

function statusBadgeClass(isActive: boolean) {
  if (isActive) {
    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
  }
  return 'border-red-500/30 bg-red-500/10 text-red-700 dark:border-red-500/40 dark:bg-red-500/20 dark:text-red-300'
}

function formatDay(value?: number | null) {
  if (!value) return '-'
  return String(value).padStart(2, '0')
}

function formatBankName(value: string) {
  const trimmed = value?.trim()
  return trimmed ? trimmed : '-'
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold">Meios de pagamento</h2>
          <p class="text-sm text-muted-foreground">Cadastre cartões e outros meios usados nos lançamentos.</p>
        </div>
        <Button @click="emit('create')">Novo meio</Button>
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

    <div v-else-if="!props.methods.length" class="px-6 py-10 text-center">
      <p class="text-sm text-muted-foreground">Sem meios de pagamento cadastrados.</p>
      <Button class="mt-4" @click="emit('create')">Criar primeiro meio</Button>
    </div>

    <div v-else class="px-2 py-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nome</TableHead>
            <TableHead>Tipo</TableHead>
            <TableHead>Banco</TableHead>
            <TableHead>Fechamento</TableHead>
            <TableHead>Vencimento</TableHead>
            <TableHead>Status</TableHead>
            <TableHead class="text-right">Ações</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="method in props.methods" :key="method.id">
            <TableCell class="font-medium">{{ method.name }}</TableCell>
            <TableCell>
              <Badge variant="outline" :class="kindBadgeClass(method.kind)">
                {{ kindLabel(method.kind) }}
              </Badge>
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ formatBankName(method.bank_name) }}
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ formatDay(method.closing_day) }}
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ formatDay(method.due_day) }}
            </TableCell>
            <TableCell>
              <Badge variant="outline" :class="statusBadgeClass(method.is_active)">
                {{ method.is_active ? 'Ativo' : 'Inativo' }}
              </Badge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-2">
                <Button
                  variant="secondary"
                  size="sm"
                  :disabled="props.actionsDisabled"
                  :title="props.actionsDisabled ? props.actionsDisabledReason : undefined"
                  @click="emit('edit', method)"
                >
                  Editar
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  :disabled="props.actionsDisabled"
                  :title="props.actionsDisabled ? props.actionsDisabledReason : undefined"
                  @click="emit('remove', method)"
                >
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
