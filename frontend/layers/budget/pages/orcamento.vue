<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import BudgetTable from '../components/BudgetTable.vue'
import type { BudgetItem, CategoryOption } from '../types/budget'
import { useBudgetService } from '../services/budget'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'
import { Label } from '~/layers/shared/components/ui/label'
import MonthPicker from '../components/MonthPicker.vue'

interface BudgetRow {
  category_id: number
  category_name: string
  planned_amount: number
  actual_amount: number
  target_percent: number
  inactive: boolean
}

definePageMeta({
  layout: 'default',
})

const { getSummary, setItemsBulk, listCategories } = useBudgetService()

const items = ref<BudgetItem[]>([])
const totalIncome = ref(0)
const categories = ref<CategoryOption[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)
const categoriesLoading = ref(false)
const categoriesError = ref<string | null>(null)

const saving = ref(false)
const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const monthValue = ref(getCurrentMonthValue())

const monthParam = computed(() => (monthValue.value ? `${monthValue.value}-01` : ''))

const monthLabel = computed(() => {
  if (!monthValue.value) return ''
  const [year, month] = monthValue.value.split('-')
  if (!year || !month) return ''
  return new Date(Number(year), Number(month) - 1, 1).toLocaleDateString('pt-BR', {
    month: 'long',
    year: 'numeric',
  })
})

const eligibleCategories = computed(() =>
  categories.value.filter((category) => category.direction === 'OUT' && category.is_budget_relevant),
)

const incomeCategories = computed(() =>
  categories.value.filter((category) => category.direction === 'IN' && category.is_budget_relevant),
)

const draftPercents = ref<Record<number, number>>({})

function isInactiveForMonth(category: CategoryOption) {
  if (!category.inactive_from_month || !monthParam.value) return false
  const inactiveDate = new Date(category.inactive_from_month)
  const monthDate = new Date(monthParam.value)
  if (Number.isNaN(inactiveDate.getTime()) || Number.isNaN(monthDate.getTime())) return false
  return inactiveDate.getTime() <= monthDate.getTime()
}

const rowData = computed<BudgetRow[]>(() => {
  const itemMap = new Map(items.value.map((item) => [item.category_id, item]))
  return eligibleCategories.value.map((category) => {
    const item = itemMap.get(category.id)
    const percent = draftPercents.value[category.id] ?? item?.target_percent ?? 0
    const planned = totalIncome.value * (percent / 100)
    return {
      category_id: category.id,
      category_name: category.name,
      planned_amount: planned,
      actual_amount: item?.actual_amount ?? 0,
      target_percent: percent,
      inactive: isInactiveForMonth(category),
    }
  })
})

const totalPercent = computed(() =>
  rowData.value.reduce((sum, row) => sum + row.target_percent, 0),
)

const totalPercentDisplay = computed(() => Math.round(totalPercent.value))

const saveState = computed(() => (totalPercentDisplay.value === 100 ? 'ready' : 'adjust'))

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

async function fetchSummary(options: { silent?: boolean } = {}) {
  if (!monthParam.value) return

  const showLoading = !options.silent && items.value.length === 0
  if (showLoading) {
    loading.value = true
  }
  if (!options.silent) {
    loadError.value = null
  }

  try {
    const summary = await getSummary(monthParam.value)
    items.value = summary.items
    totalIncome.value = summary.total_income || 0
    resetDraftPercents()
  } catch (error) {
    const message = getApiErrorMessage(error)
    if (options.silent) {
      feedback.value = { type: 'error', message }
    } else {
      loadError.value = message
    }
  } finally {
    loading.value = false
  }
}

async function fetchCategories() {
  categoriesLoading.value = true
  categoriesError.value = null
  try {
    categories.value = await listCategories(true, monthParam.value)
    resetDraftPercents()
  } catch (error) {
    categoriesError.value = getApiErrorMessage(error)
  } finally {
    categoriesLoading.value = false
  }
}

async function refreshData() {
  await Promise.all([fetchSummary(), fetchCategories()])
}

function resetDraftPercents() {
  if (!eligibleCategories.value.length) {
    draftPercents.value = {}
    return
  }

  const itemMap = new Map(items.value.map((item) => [item.category_id, item]))
  const next: Record<number, number> = {}
  for (const category of eligibleCategories.value) {
    const rawPercent = itemMap.get(category.id)?.target_percent ?? 0
    next[category.id] = isInactiveForMonth(category) ? 0 : Math.round(rawPercent)
  }
  draftPercents.value = next
}

function updatePercent(payload: { categoryId: number; value: number }) {
  draftPercents.value = {
    ...draftPercents.value,
    [payload.categoryId]: payload.value,
  }
}

async function saveBudget() {
  if (!monthParam.value) {
    feedback.value = { type: 'error', message: 'Selecione um mês válido.' }
    return
  }
  if (totalPercentDisplay.value !== 100) {
    feedback.value = { type: 'error', message: 'A soma dos percentuais deve fechar em 100%.' }
    return
  }

  saving.value = true
  feedback.value = null

  try {
    const payloads = rowData.value.map((row) => ({
      category_id: row.category_id,
      target_percent: row.target_percent,
    }))
    await setItemsBulk(monthParam.value, payloads)

    await fetchSummary({ silent: true })
    feedback.value = { type: 'success', message: 'Orçamento salvo com sucesso.' }
  } catch (error) {
    feedback.value = { type: 'error', message: getApiErrorMessage(error) }
  } finally {
    saving.value = false
  }
}

watch(monthValue, () => {
  items.value = []
  totalIncome.value = 0
  refreshData()
})

onMounted(async () => {
  await refreshData()
})

function getCurrentMonthValue() {
  const now = new Date()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  return `${now.getFullYear()}-${month}`
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Orçamento</h1>
        <p class="text-sm text-muted-foreground">
          Distribuição percentual das entradas {{ monthLabel ? `para ${monthLabel}` : '' }}.
        </p>
      </div>
      <div class="flex w-full items-end gap-3 sm:w-auto sm:flex-wrap">
        <div class="flex-[3] flex flex-col gap-1 sm:flex-none">
          <Label for="budget-month" class="text-xs text-muted-foreground">Mês de referência</Label>
          <MonthPicker id="budget-month" v-model="monthValue" :disabled="loading || categoriesLoading" />
        </div>
        <Button
          variant="outline"
          class="flex-[2] w-full sm:flex-none sm:w-auto"
          :disabled="loading || categoriesLoading"
          @click="refreshData"
        >
          Atualizar
        </Button>
      </div>
    </div>

    <div v-if="feedback" class="rounded-md border px-4 py-3 text-sm" :class="feedback.type === 'error'
      ? 'border-destructive/30 bg-destructive/10 text-destructive'
      : 'border-primary/30 bg-primary/10 text-primary'">
      <div class="flex items-center justify-between gap-4">
        <span>{{ feedback.message }}</span>
        <Button variant="ghost" size="sm" @click="feedback = null">Fechar</Button>
      </div>
    </div>

    <div v-if="categoriesError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ categoriesError }}
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <div class="rounded-lg border bg-card p-4">
        <p class="text-xs text-muted-foreground">Total de entradas relevantes</p>
        <p class="mt-2 text-lg font-semibold">{{ formatCurrency(totalIncome) }}</p>
        <div class="mt-3 flex flex-wrap gap-2">
          <span
            v-for="category in incomeCategories"
            :key="category.id"
            class="rounded-full border border-sky-500/30 bg-sky-500/10 px-2 py-1 text-xs text-sky-700 dark:border-sky-500/40 dark:bg-sky-500/20 dark:text-sky-300"
          >
            {{ category.name }}
          </span>
          <span v-if="!incomeCategories.length" class="text-xs text-muted-foreground">
            Nenhuma categoria de entrada relevante.
          </span>
        </div>
      </div>
      <div class="rounded-lg border bg-card p-4">
        <p class="text-xs text-muted-foreground">Percentual distribuído</p>
        <p class="mt-2 text-lg font-semibold">
          {{ totalPercentDisplay }}%
          <span v-if="totalPercentDisplay === 100" class="ml-2 text-xs text-emerald-600 dark:text-emerald-400">OK</span>
          <span v-else class="ml-2 text-xs text-amber-600 dark:text-amber-400">Ajustar</span>
        </p>
        <p class="mt-1 text-xs text-muted-foreground">Precisa fechar em 100% para salvar.</p>
      </div>
      <div class="rounded-lg border bg-card p-4">
        <p class="text-xs text-muted-foreground">Total planejado</p>
        <p class="mt-2 text-lg font-semibold">
          {{ formatCurrency(totalIncome * (totalPercentDisplay / 100)) }}
        </p>
      </div>
    </div>

    <BudgetTable
      :rows="rowData"
      :loading="loading || categoriesLoading"
      :error="loadError"
      :save-state="saveState"
      :saving="saving"
      @update-percent="updatePercent"
      @save="saveBudget"
      @retry="fetchSummary"
    />
  </div>
</template>
