<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CaseTable from '../../../components/CaseTable.vue'
import CaseFormSheet from '../../../components/CaseFormSheet.vue'
import CaseInstallmentsSheet from '../../../components/CaseInstallmentsSheet.vue'
import CaseDeleteDialog from '../../../components/CaseDeleteDialog.vue'
import type { PaymentMethod, Person, PicuinhaCase, PicuinhaCaseInstallment } from '../../../types/picuinha'
import type { Category } from '~/layers/categories/types/category'
import { usePicuinhasService } from '../../../services/picuinhas'
import { useInstallmentsService } from '~/layers/payment-methods/services/installments'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'

definePageMeta({
  layout: 'default',
})

const route = useRoute()
const router = useRouter()
const personId = Number(route.params.id)

const { listPersons, listCases, createCase, deleteCase, listCaseInstallments, updateCaseInstallment, listPaymentMethods, listCategories } =
  usePicuinhasService()
const { createInstallment } = useInstallmentsService()

const person = ref<Person | null>(null)
const cases = ref<PicuinhaCase[]>([])
const casesLoading = ref(true)
const casesError = ref<string | null>(null)

const paymentMethods = ref<PaymentMethod[]>([])
const categories = ref<Category[]>([])
const formOpen = ref(false)
const formSubmitting = ref(false)
const formError = ref<string | null>(null)

const installmentsOpen = ref(false)
const installmentsCase = ref<PicuinhaCase | null>(null)
const installments = ref<PicuinhaCaseInstallment[]>([])
const installmentsLoading = ref(false)
const installmentsError = ref<string | null>(null)
const updatingInstallmentId = ref<number | null>(null)

const deleteOpen = ref(false)
const deleteTarget = ref<PicuinhaCase | null>(null)
const deleting = ref(false)

const pageSubtitle = computed(() => (person.value ? `Acompanhe as picuinhas de ${person.value.name}.` : ''))

async function fetchPerson() {
  if (!Number.isFinite(personId)) {
    throw new Error('Pessoa inválida.')
  }
  const people = await listPersons()
  person.value = people.find((item) => item.id === personId) || null
  if (!person.value) {
    throw new Error('Pessoa não encontrada.')
  }
}

async function fetchCases() {
  casesLoading.value = true
  casesError.value = null
  try {
    cases.value = await listCases(personId)
  } catch (error) {
    casesError.value = getApiErrorMessage(error)
  } finally {
    casesLoading.value = false
  }
}

async function fetchReferenceData() {
  try {
    paymentMethods.value = await listPaymentMethods()
    categories.value = await listCategories()
  } catch (error) {
    casesError.value = getApiErrorMessage(error)
  }
}

async function refreshAll() {
  try {
    await fetchPerson()
    await fetchCases()
    await fetchReferenceData()
  } catch (error) {
    casesError.value = getApiErrorMessage(error)
  }
}

function openCreate() {
  formError.value = null
  formOpen.value = true
}

async function handleCreate(payload: {
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
  if (!person.value) return
  formSubmitting.value = true
  formError.value = null
  try {
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
        person_id: person.value.id,
        title: payload.title,
        case_type: 'CARD_INSTALLMENT',
        total_amount: plan.total_amount,
        installment_count: plan.installment_count,
        installment_amount: plan.installment_amount,
        start_date: plan.start_month,
        payment_method_id: plan.payment_method_id,
        installment_plan_id: plan.id,
        category_id: payload.category_id,
        interest_rate: payload.interest_rate,
        interest_rate_unit: payload.interest_rate_unit,
        recurrence_interval_months: payload.recurrence_interval_months,
      })
    } else {
      await createCase({
        person_id: person.value.id,
        title: payload.title,
        case_type: payload.case_type,
        total_amount: payload.total_amount,
        installment_count: payload.installment_count,
        installment_amount: payload.installment_amount,
        start_date: payload.start_date || '',
        interest_rate: payload.interest_rate,
        interest_rate_unit: payload.interest_rate_unit,
        recurrence_interval_months: payload.recurrence_interval_months,
      })
    }
    formOpen.value = false
    await fetchCases()
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    formSubmitting.value = false
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
    deleteOpen.value = false
  } catch (error) {
    casesError.value = getApiErrorMessage(error)
  } finally {
    deleting.value = false
  }
}

function goBack() {
  router.push('/picuinhas/pessoas')
}

onMounted(refreshAll)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <Button variant="ghost" class="px-0" @click="goBack">
          ← Voltar para pessoas
        </Button>
        <h1 class="text-2xl font-semibold tracking-tight">
          {{ person ? `Picuinhas · ${person.name}` : 'Picuinhas' }}
        </h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" :disabled="casesLoading" @click="fetchCases">Atualizar</Button>
    </div>

    <CaseTable
      :cases="cases"
      :loading="casesLoading"
      :error="casesError"
      @create="openCreate"
      @viewInstallments="openInstallments"
      @remove="requestDelete"
      @retry="fetchCases"
    />

    <CaseFormSheet
      v-model:open="formOpen"
      :submitting="formSubmitting"
      :error-message="formError"
      :payment-methods="paymentMethods"
      :categories="categories"
      @submit="handleCreate"
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
