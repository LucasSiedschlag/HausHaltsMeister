<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import InstallmentFormCard from '../../components/InstallmentFormCard.vue'
import type { InstallmentCategoryOption } from '../../types/installment'
import type { PaymentMethod } from '../../types/payment-method'
import { useInstallmentsService } from '../../services/installments'
import type { Person } from '~/layers/picuinhas/types/picuinha'
import { usePicuinhasService } from '~/layers/picuinhas/services/picuinhas'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'

definePageMeta({
  layout: 'default',
})

const { listPaymentMethods, listCategories, createInstallment } = useInstallmentsService()
const { listPersons, createCase } = usePicuinhasService()

const categories = ref<InstallmentCategoryOption[]>([])
const paymentMethods = ref<PaymentMethod[]>([])
const persons = ref<Person[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)

const categoriesLoading = ref(false)
const paymentMethodsLoading = ref(false)

const formError = ref<string | null>(null)
const submitting = ref(false)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const pageSubtitle = computed(() => 'Registre compras no cartão e gere parcelas automáticas.')

async function fetchCategories() {
  categoriesLoading.value = true
  try {
    categories.value = await listCategories()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    categoriesLoading.value = false
  }
}

async function fetchPaymentMethods() {
  paymentMethodsLoading.value = true
  try {
    paymentMethods.value = await listPaymentMethods()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    paymentMethodsLoading.value = false
  }
}

async function fetchPersons() {
  try {
    persons.value = await listPersons()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  }
}

async function refreshBaseData() {
  loading.value = true
  loadError.value = null
  await Promise.all([fetchCategories(), fetchPaymentMethods(), fetchPersons()])
  loading.value = false
}

async function handleCreate(payload: {
  description: string
  amount_mode: 'TOTAL' | 'INSTALLMENT'
  total_amount?: number
  installment_amount?: number
  count: number
  category_id: number
  payment_method_id: number
  purchase_date: string
  person_id?: number
}) {
  submitting.value = true
  formError.value = null
  try {
    const { person_id, ...installmentPayload } = payload
    const plan = await createInstallment(installmentPayload)
    if (person_id) {
      await createCase({
        person_id,
        title: payload.description,
        case_type: 'CARD_INSTALLMENT',
        total_amount: plan.total_amount,
        installment_count: plan.installment_count,
        installment_amount: plan.installment_amount,
        start_date: plan.start_month,
        payment_method_id: plan.payment_method_id,
        installment_plan_id: plan.id,
        category_id: payload.category_id,
      })
    }
    feedback.value = { type: 'success', message: 'Parcelamento registrado com sucesso.' }
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

onMounted(refreshBaseData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Parcelamentos · Nova compra</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" :disabled="loading" @click="refreshBaseData">Atualizar dados</Button>
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

    <InstallmentFormCard
      :categories="categories"
      :payment-methods="paymentMethods"
      :persons="persons"
      :categories-loading="categoriesLoading"
      :payment-methods-loading="paymentMethodsLoading"
      :submitting="submitting"
      :error-message="formError"
      @submit="handleCreate"
    />
  </div>
</template>
