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
import type { Person } from '../types/picuinha'

interface Props {
  persons: Person[]
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
  edit: [person: Person]
  remove: [person: Person]
  view: [person: Person]
  retry: []
}>()

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

function balanceBadgeClass(balance: number) {
  if (balance > 0) {
    return 'border-red-500/30 bg-red-500/10 text-red-700 dark:border-red-500/40 dark:bg-red-500/20 dark:text-red-300'
  }
  if (balance < 0) {
    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
  }
  return 'border-muted/50 bg-muted/40 text-muted-foreground'
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold">Pessoas</h2>
          <p class="text-sm text-muted-foreground">Cadastre pessoas e acompanhe o valor em aberto.</p>
        </div>
        <Button @click="emit('create')">Nova pessoa</Button>
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

    <div v-else-if="!props.persons.length" class="px-6 py-10 text-center">
      <p class="text-sm text-muted-foreground">Sem pessoas cadastradas.</p>
      <Button class="mt-4" @click="emit('create')">Criar primeira pessoa</Button>
    </div>

    <div v-else class="px-2 py-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nome</TableHead>
            <TableHead>Observações</TableHead>
            <TableHead>Em aberto</TableHead>
            <TableHead class="text-right">Ações</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="person in props.persons" :key="person.id">
            <TableCell class="font-medium">{{ person.name }}</TableCell>
            <TableCell class="text-sm text-muted-foreground">{{ person.notes || '-' }}</TableCell>
            <TableCell>
              <Badge variant="outline" :class="balanceBadgeClass(person.balance)">
                {{ formatCurrency(person.balance) }}
              </Badge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  :disabled="props.actionsDisabled"
                  :title="props.actionsDisabled ? props.actionsDisabledReason : undefined"
                  @click="emit('view', person)"
                >
                  Picuinhas
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  :disabled="props.actionsDisabled"
                  :title="props.actionsDisabled ? props.actionsDisabledReason : undefined"
                  @click="emit('edit', person)"
                >
                  Editar
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  :disabled="props.actionsDisabled"
                  :title="props.actionsDisabled ? props.actionsDisabledReason : undefined"
                  @click="emit('remove', person)"
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
