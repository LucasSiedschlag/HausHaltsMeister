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
import type { PicuinhaEntry } from '../types/picuinha'

interface EntryRow extends PicuinhaEntry {
  person_name?: string
  payment_method_name?: string
}

interface Props {
  entries: EntryRow[]
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
  edit: [entry: EntryRow]
  remove: [entry: EntryRow]
  retry: []
}>()

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

function kindLabel(kind: string) {
  return kind === 'PLUS' ? 'Emprestei' : 'Recebi'
}

function kindBadgeClass(kind: string) {
  if (kind === 'PLUS') {
    return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:border-amber-500/40 dark:bg-amber-500/20 dark:text-amber-300'
  }
  return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold">Lançamentos</h2>
          <p class="text-sm text-muted-foreground">Registre empréstimos e pagamentos.</p>
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
      <p class="text-sm text-muted-foreground">Sem lançamentos cadastrados.</p>
      <Button class="mt-4" @click="emit('create')">Criar primeiro lançamento</Button>
    </div>

    <div v-else class="px-2 py-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Pessoa</TableHead>
            <TableHead>Cartão</TableHead>
            <TableHead>Tipo</TableHead>
            <TableHead>Valor</TableHead>
            <TableHead>Data</TableHead>
            <TableHead class="text-right">Ações</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="entry in props.entries" :key="entry.id">
            <TableCell class="font-medium">{{ entry.person_name || `#${entry.person_id}` }}</TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ entry.payment_method_name || '-' }}
            </TableCell>
            <TableCell>
              <Badge variant="outline" :class="kindBadgeClass(entry.kind)">
                {{ kindLabel(entry.kind) }}
              </Badge>
            </TableCell>
            <TableCell>{{ formatCurrency(entry.amount) }}</TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ new Date(entry.created_at).toLocaleDateString('pt-BR') }}
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-2">
                <Button
                  variant="secondary"
                  size="sm"
                  :disabled="props.actionsDisabled"
                  :title="props.actionsDisabled ? props.actionsDisabledReason : undefined"
                  @click="emit('edit', entry)"
                >
                  Editar
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  :disabled="props.actionsDisabled"
                  :title="props.actionsDisabled ? props.actionsDisabledReason : undefined"
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
