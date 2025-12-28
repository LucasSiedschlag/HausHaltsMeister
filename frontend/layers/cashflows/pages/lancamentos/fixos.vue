<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import CashflowTable from '../../components/CashflowTable.vue'
import CashflowFormSheet from '../../components/CashflowFormSheet.vue'
import CashflowConfirmDialog from '../../components/CashflowConfirmDialog.vue'
import MonthPicker from '../../components/MonthPicker.vue'
import type { CashflowEntry, CreateCashflowRequest } from '../../types/cashflow'
import type { Category } from '~/layers/categories/types/category'
import type { PaymentMethod } from '~/layers/payment-methods/types/payment-method'
import { useCashflowsService } from '../../services/cashflows'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'
import { Label } from '~/layers/shared/components/ui/label'

definePageMeta({
  layout: 'default',
})

const {
  listCashflows,
  createCashflow,
  updateCashflow,
  deleteCashflow,
  reverseCashflow,
  copyFixedCashflows,
  listCategories,
  listPaymentMethods,
} = useCashflowsService()

const entries = ref<CashflowEntry[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)

const categories = ref<Category[]>([])
const categoriesLoading = ref(false)
const categoriesError = ref<string | null>(null)

const paymentMethods = ref<PaymentMethod[]>([])
const paymentMethodsLoading = ref(false)
const paymentMethodsError = ref<string | null>(null)

const monthValue = ref(getCurrentMonthValue())
const monthParam = computed(() => (monthValue.value ? `${monthValue.value}-01` : ''))

const pageSubtitle = computed(() => 'Controle gastos fixos replicados automaticamente nos meses seguintes.')

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const selectedEntry = ref<CashflowEntry | null>(null)
const formError = ref<string | null>(null)
const submitting = ref(false)

const confirmOpen = ref(false)
const confirmMode = ref<'delete' | 'reverse'>('delete')
const confirmTarget = ref<CashflowEntry | null>(null)
const confirming = ref(false)
const copying = ref(false)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const fixedCategory = computed(() =>
  categories.value.find((category) =>
    category.direction === 'OUT' && category.name.trim().toLowerCase() === 'custos fixos',
  ) || null,
)

const categoryOptions = computed(() => (fixedCategory.value ? [fixedCategory.value] : []))

const paymentOptions = computed(() => paymentMethods.value.filter((method) => method.is_active))

const referenceError = computed(() => categoriesError.value || paymentMethodsError.value)

const missingFixedCategory = computed(() => !categoriesLoading.value && !fixedCategory.value)

async function fetchEntries() {
  if (!monthParam.value) return
  loading.value = true
  loadError.value = null
  try {
    entries.value = await listCashflows(monthParam.value, { direction: 'OUT', is_fixed: true })
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function fetchCategories() {
  categoriesLoading.value = true
  categoriesError.value = null
  try {
    categories.value = await listCategories(true, monthParam.value)
  } catch (error) {
    categoriesError.value = getApiErrorMessage(error)
  } finally {
    categoriesLoading.value = false
  }
}

async function fetchPaymentMethods() {
  paymentMethodsLoading.value = true
  paymentMethodsError.value = null
  try {
    paymentMethods.value = await listPaymentMethods()
  } catch (error) {
    paymentMethodsError.value = getApiErrorMessage(error)
  } finally {
    paymentMethodsLoading.value = false
  }
}

async function refreshData() {
  await Promise.all([fetchCategories(), fetchPaymentMethods()])
  await fetchEntries()
  await ensureFixedCopied()
}

function openCreate() {
  if (categoriesLoading.value || paymentMethodsLoading.value) {
    feedback.value = { type: 'error', message: 'Aguarde o carregamento dos dados de referência.' }
    return
  }
  if (referenceError.value) {
    feedback.value = { type: 'error', message: referenceError.value }
    return
  }
  if (!fixedCategory.value) {
    feedback.value = { type: 'error', message: 'Categoria "Custos fixos" não encontrada.' }
    return
  }
  formMode.value = 'create'
  selectedEntry.value = null
  formError.value = null
  formOpen.value = true
}

function openEdit(entry: CashflowEntry) {
  formMode.value = 'edit'
  selectedEntry.value = entry
  formError.value = null
  formOpen.value = true
}

async function handleSubmit(payload: CreateCashflowRequest) {
  submitting.value = true
  formError.value = null
  try {
    if (formMode.value === 'edit' && selectedEntry.value) {
      const updated = await updateCashflow(selectedEntry.value.id, payload)
      entries.value = entries.value.map((item) => (item.id === updated.id ? updated : item))
      feedback.value = { type: 'success', message: 'Lançamento atualizado com sucesso.' }
    } else {
      const created = await createCashflow(payload)
      entries.value = [created, ...entries.value]
      feedback.value = { type: 'success', message: 'Lançamento fixo criado com sucesso.' }
    }
    formOpen.value = false
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function requestDelete(entry: CashflowEntry) {
  confirmMode.value = 'delete'
  confirmTarget.value = entry
  confirmOpen.value = true
}

function requestReverse(entry: CashflowEntry) {
  confirmMode.value = 'reverse'
  confirmTarget.value = entry
  confirmOpen.value = true
}

async function confirmAction() {
  if (!confirmTarget.value) return
  confirming.value = true
  try {
    if (confirmMode.value === 'delete') {
      await deleteCashflow(confirmTarget.value.id)
      feedback.value = { type: 'success', message: 'Lançamento excluído com sucesso.' }
    } else {
      await reverseCashflow(confirmTarget.value.id)
      feedback.value = { type: 'success', message: 'Estorno criado com sucesso.' }
    }
    await fetchEntries()
    confirmOpen.value = false
  } catch (error) {
    feedback.value = { type: 'error', message: getApiErrorMessage(error) }
  } finally {
    confirming.value = false
  }
}

async function ensureFixedCopied() {
  if (!monthParam.value || copying.value) return
  if (entries.value.length > 0) return
  const previousMonth = getPreviousMonthValue(monthValue.value)
  if (!previousMonth) return

  copying.value = true
  try {
    const result = await copyFixedCashflows(`${previousMonth}-01`, monthParam.value)
    if (result.copied_count > 0) {
      feedback.value = {
        type: 'success',
        message: `Fixos copiados do mês anterior (${result.copied_count}).`,
      }
    }
    await fetchEntries()
  } catch (error) {
    feedback.value = { type: 'error', message: getApiErrorMessage(error) }
  } finally {
    copying.value = false
  }
}

watch(formOpen, (open) => {
  if (!open) {
    formError.value = null
  }
})

watch(monthValue, () => {
  entries.value = []
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

function getPreviousMonthValue(value: string) {
  const [year, month] = value.split('-').map(Number)
  if (!year || !month) return ''
  const previous = new Date(year, month - 2, 1)
  const prevMonth = String(previous.getMonth() + 1).padStart(2, '0')
  return `${previous.getFullYear()}-${prevMonth}`
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Fixos</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <div class="flex flex-col gap-3 sm:flex-row sm:items-end">
        <div class="space-y-1">
          <Label for="month-fixos">Mês de referência</Label>
          <MonthPicker id="month-fixos" v-model="monthValue" :disabled="loading" />
        </div>
        <Button variant="outline" :disabled="loading" @click="refreshData">Atualizar</Button>
        <Button :disabled="categoriesLoading || paymentMethodsLoading" @click="openCreate">
          Novo lançamento
        </Button>
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

    <div v-if="referenceError" class="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      {{ referenceError }}
    </div>

    <div v-if="missingFixedCategory" class="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      Categoria "Custos fixos" não encontrada. Atualize o cadastro de categorias.
    </div>

    <CashflowTable
      :entries="entries"
      :loading="loading"
      :error="loadError"
      title="Fixos do mês"
      description="Lançamentos fixos com categoria Custos fixos."
      @create="openCreate"
      @edit="openEdit"
      @remove="requestDelete"
      @reverse="requestReverse"
      @retry="fetchEntries"
    />

    <CashflowFormSheet
      v-model:open="formOpen"
      :mode="formMode"
      :entry="selectedEntry"
      :categories="categoryOptions"
      :payment-methods="paymentOptions"
      :direction="'OUT'"
      :is-fixed="true"
      :reference-date="monthParam"
      :category-locked="true"
      :submitting="submitting"
      :error-message="formError"
      @submit="handleSubmit"
    />

    <CashflowConfirmDialog
      v-model:open="confirmOpen"
      :mode="confirmMode"
      :entry="confirmTarget"
      :submitting="confirming"
      @confirm="confirmAction"
    />
  </div>
</template>
