<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import CaseTable from '../../components/CaseTable.vue'
import CaseFormSheet from '../../components/CaseFormSheet.vue'
import CaseInstallmentsSheet from '../../components/CaseInstallmentsSheet.vue'
import CaseDeleteDialog from '../../components/CaseDeleteDialog.vue'
import type { PaymentMethod, Person, PicuinhaCase, PicuinhaCaseInstallment } from '../../types/picuinha'
import type { Category } from '~/layers/categories/types/category'
import { usePicuinhasService } from '../../services/picuinhas'
import { useInstallmentsService } from '~/layers/payment-methods/services/installments'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'

definePageMeta({
  layout: 'default',
})

const {
  listPersons,
  listCases,
  createCase,
  deleteCase,
  listCaseInstallments,
  updateCaseInstallment,
  listPaymentMethods,
  listCategories,
} = usePicuinhasService()
const { createInstallment } = useInstallmentsService()

const cases = ref<PicuinhaCase[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)

const persons = ref<Person[]>([])
const personsLoading = ref(false)
const personsError = ref<string | null>(null)

const paymentMethods = ref<PaymentMethod[]>([])
const paymentMethodsLoading = ref(false)
const paymentMethodsError = ref<string | null>(null)

const categories = ref<Category[]>([])
const categoriesLoading = ref(false)
const categoriesError = ref<string | null>(null)

const route = useRoute()
const selectedPersonId = ref<string>('')

const formOpen = ref(false)
const formError = ref<string | null>(null)
const submitting = ref(false)

const installmentsOpen = ref(false)
const installmentsCase = ref<PicuinhaCase | null>(null)
const installments = ref<PicuinhaCaseInstallment[]>([])
const installmentsLoading = ref(false)
const installmentsError = ref<string | null>(null)
const updatingInstallmentId = ref<number | null>(null)

const deleteOpen = ref(false)
const deleteTarget = ref<PicuinhaCase | null>(null)
const deleting = ref(false)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const pageSubtitle = computed(() =>
  selectedPersonId.value ? 'Acompanhe as picuinhas da pessoa selecionada.' : 'Visão geral de todas as picuinhas.',
)

const canCreate = computed(() => Boolean(selectedPersonId.value))

async function fetchPersons() {
  personsLoading.value = true
  personsError.value = null
  try {
    persons.value = await listPersons()
  } catch (error) {
    personsError.value = getApiErrorMessage(error)
  } finally {
    personsLoading.value = false
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

async function fetchCategories() {
  categoriesLoading.value = true
  categoriesError.value = null
  try {
    categories.value = await listCategories()
  } catch (error) {
    categoriesError.value = getApiErrorMessage(error)
  } finally {
    categoriesLoading.value = false
  }
}

async function fetchCases() {
  loading.value = true
  loadError.value = null
  try {
    if (selectedPersonId.value) {
      cases.value = await listCases(Number(selectedPersonId.value))
      return
    }
    const people = persons.value.length ? persons.value : await listPersons()
    persons.value = people
    const results = await Promise.all(
      people.map((person) =>
        listCases(person.id).then((items) =>
          items.map((item) => ({
            ...item,
            person_id: person.id,
            person_name: person.name,
          })),
        ),
      ),
    )
    cases.value = results.flat()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function refreshData() {
  await Promise.all([fetchPersons(), fetchCases(), fetchPaymentMethods(), fetchCategories()])
}

function openCreate() {
  if (!selectedPersonId.value) {
    feedback.value = { type: 'error', message: 'Selecione uma pessoa para criar a picuinha.' }
    return
  }
  formError.value = null
  formOpen.value = true
}

async function handleSubmit(payload: {
  title: string
  case_type: PicuinhaCase['case_type']
  amount_mode: 'TOTAL' | 'INSTALLMENT'
  total_amount?: number
  installment_amount?: number
  installment_count?: number
  start_date?: string
  purchase_date?: string
  payment_method_id?: number
  category_id?: number
  interest_rate?: number
  interest_rate_unit?: string
  recurrence_interval_months?: number
}) {
  if (!selectedPersonId.value) {
    formError.value = 'Selecione uma pessoa para continuar.'
    return
  }
  submitting.value = true
  formError.value = null
  try {
    const resolvedStartDate = payload.start_date || payload.purchase_date
    if (payload.case_type === 'CARD_INSTALLMENT') {
      if (!payload.purchase_date || !payload.payment_method_id || !payload.category_id || !payload.installment_count) {
        throw new Error('Dados incompletos para compra no cartão.')
      }
      const plan = await createInstallment({
        description: payload.title,
        amount_mode: payload.amount_mode,
        total_amount: payload.total_amount,
        installment_amount: payload.installment_amount,
        count: payload.installment_count,
        category_id: payload.category_id,
        payment_method_id: payload.payment_method_id,
        purchase_date: payload.purchase_date,
      })
      await createCase({
        person_id: Number(selectedPersonId.value),
        title: payload.title,
        case_type: 'CARD_INSTALLMENT',
        total_amount: plan.total_amount,
        installment_count: plan.installment_count,
        installment_amount: plan.installment_amount,
        start_date: plan.start_month || resolvedStartDate || '',
        payment_method_id: plan.payment_method_id,
        installment_plan_id: plan.id,
        category_id: payload.category_id,
        interest_rate: payload.interest_rate,
        interest_rate_unit: payload.interest_rate_unit,
        recurrence_interval_months: payload.recurrence_interval_months,
      })
    } else {
      if (!resolvedStartDate) {
        throw new Error('Informe a data inicial.')
      }
      await createCase({
        person_id: Number(selectedPersonId.value),
        title: payload.title,
        case_type: payload.case_type,
        total_amount: payload.total_amount,
        installment_count: payload.installment_count,
        installment_amount: payload.installment_amount,
        start_date: resolvedStartDate,
        interest_rate: payload.interest_rate,
        interest_rate_unit: payload.interest_rate_unit,
        recurrence_interval_months: payload.recurrence_interval_months,
      })
    }
    feedback.value = { type: 'success', message: 'Picuinha criada com sucesso.' }
    await fetchCases()
    formOpen.value = false
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function openInstallments(picCase: PicuinhaCase) {
  installmentsCase.value = picCase
  installmentsOpen.value = true
  loadInstallments(picCase.id)
}

async function loadInstallments(caseId: number) {
  installmentsLoading.value = true
  installmentsError.value = null
  try {
    installments.value = await listCaseInstallments(caseId)
  } catch (error) {
    installmentsError.value = getApiErrorMessage(error)
  } finally {
    installmentsLoading.value = false
  }
}

async function handleUpdateInstallment(payload: { id: number; is_paid: boolean; extra_amount: number }) {
  updatingInstallmentId.value = payload.id
  try {
    const updated = await updateCaseInstallment(payload.id, payload)
    installments.value = installments.value.map((item) => (item.id === updated.id ? updated : item))
    await fetchCases()
  } catch (error) {
    installmentsError.value = getApiErrorMessage(error)
  } finally {
    updatingInstallmentId.value = null
  }
}

function requestDelete(picCase: PicuinhaCase) {
  deleteTarget.value = picCase
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteCase(deleteTarget.value.id)
    cases.value = cases.value.filter((item) => item.id !== deleteTarget.value?.id)
    feedback.value = { type: 'success', message: 'Picuinha excluída com sucesso.' }
    deleteOpen.value = false
  } catch (error) {
    feedback.value = { type: 'error', message: getApiErrorMessage(error) }
  } finally {
    deleting.value = false
  }
}

watch(formOpen, (open) => {
  if (!open) {
    formError.value = null
  }
})

watch(selectedPersonId, () => {
  fetchCases()
})

onMounted(async () => {
  const personParam = route.query.person_id
  if (typeof personParam === 'string' && personParam) {
    selectedPersonId.value = personParam
  }
  await refreshData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Picuinhas · Lançamentos</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" :disabled="personsLoading" @click="refreshData">Atualizar lista</Button>
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

    <div v-if="personsError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ personsError }}
    </div>

    <div v-if="paymentMethodsError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ paymentMethodsError }}
    </div>

    <div v-if="categoriesError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ categoriesError }}
    </div>

    <div class="rounded-lg border bg-card px-4 py-4">
      <div class="space-y-2">
        <Label for="picuinha-person">Pessoa</Label>
        <Select id="picuinha-person" v-model="selectedPersonId" :disabled="personsLoading">
          <option value="">Selecione uma pessoa</option>
          <option v-for="person in persons" :key="person.id" :value="String(person.id)">
            {{ person.name }}
          </option>
        </Select>
      </div>
    </div>

    <CaseTable
      :cases="cases"
      :loading="loading || personsLoading || paymentMethodsLoading || categoriesLoading"
      :error="loadError"
      @create="openCreate"
      @viewInstallments="openInstallments"
      @remove="requestDelete"
      @retry="fetchCases"
    />

    <CaseFormSheet
      v-model:open="formOpen"
      :submitting="submitting"
      :error-message="formError"
      :payment-methods="paymentMethods"
      :categories="categories"
      @submit="handleSubmit"
    />

    <CaseInstallmentsSheet
      v-model:open="installmentsOpen"
      :pic-case="installmentsCase"
      :installments="installments"
      :loading="installmentsLoading"
      :error="installmentsError"
      :updating-id="updatingInstallmentId"
      @updateInstallment="handleUpdateInstallment"
    />

    <CaseDeleteDialog
      v-model:open="deleteOpen"
      :pic-case="deleteTarget"
      :submitting="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>
