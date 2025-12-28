<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import MonthPicker from '../../components/MonthPicker.vue'
import { usePicuinhasMonthlyService } from '../../services/picuinhas'
import type { Person, PicuinhaCase, PicuinhaCaseInstallment } from '~/layers/picuinhas/types/picuinha'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Switch } from '~/layers/shared/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/layers/shared/components/ui/table'

definePageMeta({
  layout: 'default',
})

interface MonthlyInstallment {
  id: number
  case_id: number
  person_id: number
  person_name: string
  case_title: string
  case_type: PicuinhaCase['case_type']
  installment_number: number
  due_date: string
  amount: number
  extra_amount: number
  is_paid: boolean
  paid_at?: string | null
}

const { listPersons, listCases, listCaseInstallments, updateCaseInstallment } = usePicuinhasMonthlyService()

const persons = ref<Person[]>([])
const cases = ref<PicuinhaCase[]>([])
const installments = ref<MonthlyInstallment[]>([])

const loading = ref(true)
const loadError = ref<string | null>(null)
const updatingId = ref<number | null>(null)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const monthValue = ref(getCurrentMonthValue())
const monthParam = computed(() => (monthValue.value ? `${monthValue.value}-01` : ''))

const pageSubtitle = computed(() => 'Visão mensal das picuinhas, misturando todas as pessoas.')

const extras = ref<Record<number, string>>({})

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

function monthKey(value: string) {
  const [year, month] = value.split('-')
  if (!year || !month) return ''
  return `${year}-${month}`
}

function totalAmount(row: MonthlyInstallment) {
  const extra = Number(extras.value[row.id] ?? row.extra_amount ?? 0)
  return row.amount + (Number.isFinite(extra) ? extra : 0)
}

function syncExtras() {
  const next: Record<number, string> = {}
  installments.value.forEach((row) => {
    next[row.id] = String(row.extra_amount ?? 0)
  })
  extras.value = next
}

async function fetchMonthlyPicuinhas() {
  if (!monthValue.value) return
  loading.value = true
  loadError.value = null

  try {
    const people = await listPersons()
    persons.value = people

    if (!people.length) {
      cases.value = []
      installments.value = []
      return
    }

    const caseLists = await Promise.all(
      people.map((person) => listCases(person.id)),
    )

    const allCases = caseLists.flat()
    cases.value = allCases

    const casePeople = new Map<number, { person_id: number; person_name: string }>()
    allCases.forEach((picCase) => {
      const owner = people.find((person) => person.id === picCase.person_id)
      casePeople.set(picCase.id, {
        person_id: picCase.person_id,
        person_name: owner?.name || 'Pessoa',
      })
    })

    if (!allCases.length) {
      installments.value = []
      return
    }

    const rows = await Promise.all(
      allCases.map((picCase) =>
        listCaseInstallments(picCase.id).then((items) =>
          items
            .filter((installment) => monthKey(installment.due_date) === monthValue.value)
            .map((installment) => ({
              person_id: casePeople.get(picCase.id)?.person_id || picCase.person_id,
              person_name: casePeople.get(picCase.id)?.person_name || 'Pessoa',
              id: installment.id,
              case_id: installment.case_id,
              case_title: picCase.title,
              case_type: picCase.case_type,
              installment_number: installment.installment_number,
              due_date: installment.due_date,
              amount: installment.amount,
              extra_amount: installment.extra_amount,
              is_paid: installment.is_paid,
              paid_at: installment.paid_at,
            })),
        ),
      ),
    )

    const flattened = rows.flat()
    flattened.sort((a, b) => {
      const aDate = parseLocalDate(a.due_date)?.getTime() || 0
      const bDate = parseLocalDate(b.due_date)?.getTime() || 0
      if (aDate !== bDate) return aDate - bDate
      return a.person_name.localeCompare(b.person_name, 'pt-BR', { sensitivity: 'base' })
    })

    installments.value = flattened
    syncExtras()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function updateInstallment(row: MonthlyInstallment, nextPaid?: boolean) {
  const extraValue = Number(extras.value[row.id] || 0)
  if (!Number.isFinite(extraValue)) return

  updatingId.value = row.id
  try {
    const updated = await updateCaseInstallment(row.id, {
      is_paid: nextPaid ?? row.is_paid,
      extra_amount: extraValue,
    })

    installments.value = installments.value.map((item) =>
      item.id === row.id
        ? {
            ...item,
            extra_amount: updated.extra_amount,
            is_paid: updated.is_paid,
            paid_at: updated.paid_at,
          }
        : item,
    )
    extras.value = {
      ...extras.value,
      [row.id]: String(updated.extra_amount ?? 0),
    }
  } catch (error) {
    feedback.value = { type: 'error', message: getApiErrorMessage(error) }
  } finally {
    updatingId.value = null
  }
}

watch(monthValue, () => {
  installments.value = []
  fetchMonthlyPicuinhas()
})

onMounted(fetchMonthlyPicuinhas)

function getCurrentMonthValue() {
  const now = new Date()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  return `${now.getFullYear()}-${month}`
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Picuinhas do mês</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <div class="flex flex-col gap-3 sm:flex-row sm:items-end">
        <div class="space-y-1">
          <Label for="month-picuinhas">Mês de referência</Label>
          <MonthPicker id="month-picuinhas" v-model="monthValue" :disabled="loading" />
        </div>
        <Button variant="outline" :disabled="loading" @click="fetchMonthlyPicuinhas">Atualizar</Button>
      </div>
    </div>

    <div
      v-if="feedback"
      class="rounded-md border px-4 py-3 text-sm"
      :class="feedback.type === 'error' ? 'border-destructive/30 bg-destructive/10 text-destructive' : 'border-primary/30 bg-primary/10 text-primary'"
    >
      <div class="flex items-center justify-between gap-4">
        <span>{{ feedback.message }}</span>
        <Button variant="ghost" size="sm" @click="feedback = null">Fechar</Button>
      </div>
    </div>

    <div v-if="loadError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ loadError }}
    </div>

    <div v-if="loading" class="space-y-3 rounded-lg border bg-card p-6">
      <div class="h-8 w-40 animate-pulse rounded-md bg-muted" />
      <div class="space-y-3">
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
        <div class="h-10 w-full animate-pulse rounded-md bg-muted" />
      </div>
    </div>

    <div v-else-if="!installments.length" class="rounded-lg border bg-card px-6 py-10 text-center">
      <p class="text-sm text-muted-foreground">Sem picuinhas lançadas para o mês selecionado.</p>
    </div>

    <div v-else class="rounded-lg border bg-card text-card-foreground shadow-sm">
      <div class="border-b px-6 py-4">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-base font-semibold">Lançamentos do mês</h2>
            <p class="text-sm text-muted-foreground">{{ monthParam }}</p>
          </div>
        </div>
      </div>

      <div class="px-2 py-2">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Pessoa</TableHead>
              <TableHead>Picuinha</TableHead>
              <TableHead>Parcela</TableHead>
              <TableHead>Vencimento</TableHead>
              <TableHead>Valor</TableHead>
              <TableHead>Juros</TableHead>
              <TableHead>Total</TableHead>
              <TableHead class="text-right">Pago</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="row in installments" :key="row.id">
              <TableCell class="font-medium">{{ row.person_name }}</TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ row.case_title }}</TableCell>
              <TableCell>{{ row.installment_number }}</TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ formatDate(row.due_date) }}</TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ formatCurrency(row.amount) }}</TableCell>
              <TableCell>
                <Input
                  v-model="extras[row.id]"
                  type="number"
                  min="0"
                  step="0.01"
                  class="h-8 w-24"
                  :disabled="updatingId === row.id"
                  @blur="updateInstallment(row)"
                />
              </TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ formatCurrency(totalAmount(row)) }}</TableCell>
              <TableCell class="text-right">
                <Switch
                  :model-value="row.is_paid"
                  :disabled="updatingId === row.id"
                  @update:model-value="updateInstallment(row, $event)"
                />
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </div>
  </div>
</template>
