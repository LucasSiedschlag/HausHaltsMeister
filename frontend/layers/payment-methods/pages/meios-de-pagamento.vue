<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import PaymentMethodTable from '../components/PaymentMethodTable.vue'
import PaymentMethodFormSheet from '../components/PaymentMethodFormSheet.vue'
import PaymentMethodDeleteDialog from '../components/PaymentMethodDeleteDialog.vue'
import type { CreatePaymentMethodRequest, PaymentMethod, UpdatePaymentMethodRequest } from '../types/payment-method'
import { usePaymentMethodsService } from '../services/payment-methods'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'

definePageMeta({
  layout: 'default',
})

const { listPaymentMethods, createPaymentMethod, updatePaymentMethod, deletePaymentMethod } = usePaymentMethodsService()

const paymentMethods = ref<PaymentMethod[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const selectedMethod = ref<PaymentMethod | null>(null)
const formError = ref<string | null>(null)
const submitting = ref(false)

const deleteOpen = ref(false)
const deleteTarget = ref<PaymentMethod | null>(null)
const deleting = ref(false)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const pageSubtitle = computed(() => 'Gerencie cartões e outros meios usados nos lançamentos.')

async function fetchPaymentMethods() {
  loading.value = true
  loadError.value = null
  try {
    paymentMethods.value = await listPaymentMethods()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  formMode.value = 'create'
  selectedMethod.value = null
  formError.value = null
  formOpen.value = true
}

function openEdit(method: PaymentMethod) {
  formMode.value = 'edit'
  selectedMethod.value = method
  formError.value = null
  formOpen.value = true
}

async function handleSubmit(payload: CreatePaymentMethodRequest & { is_active?: boolean }) {
  submitting.value = true
  formError.value = null
  try {
    if (formMode.value === 'edit' && selectedMethod.value) {
      const updatePayload: UpdatePaymentMethodRequest = {
        name: payload.name,
        kind: payload.kind,
        bank_name: payload.bank_name,
        credit_limit: payload.credit_limit,
        closing_day: payload.closing_day,
        due_day: payload.due_day,
        is_active: payload.is_active ?? selectedMethod.value.is_active,
      }
      const updated = await updatePaymentMethod(selectedMethod.value.id, updatePayload)
      paymentMethods.value = paymentMethods.value.map((item) => (item.id === updated.id ? updated : item))
      feedback.value = { type: 'success', message: 'Meio de pagamento atualizado com sucesso.' }
    } else {
      const createPayload: CreatePaymentMethodRequest = {
        name: payload.name,
        kind: payload.kind,
        bank_name: payload.bank_name,
        credit_limit: payload.credit_limit,
        closing_day: payload.closing_day,
        due_day: payload.due_day,
      }
      const created = await createPaymentMethod(createPayload)
      paymentMethods.value = [created, ...paymentMethods.value]
      feedback.value = { type: 'success', message: 'Meio de pagamento criado com sucesso.' }
    }
    formOpen.value = false
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function requestDelete(method: PaymentMethod) {
  deleteTarget.value = method
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deletePaymentMethod(deleteTarget.value.id)
    await fetchPaymentMethods()
    feedback.value = { type: 'success', message: 'Meio de pagamento desativado com sucesso.' }
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

onMounted(fetchPaymentMethods)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Meios de Pagamento</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" @click="fetchPaymentMethods">Atualizar lista</Button>
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

    <PaymentMethodTable
      :methods="paymentMethods"
      :loading="loading"
      :error="loadError"
      @create="openCreate"
      @edit="openEdit"
      @remove="requestDelete"
      @retry="fetchPaymentMethods"
    />

    <PaymentMethodFormSheet
      v-model:open="formOpen"
      :mode="formMode"
      :method="selectedMethod"
      :submitting="submitting"
      :error-message="formError"
      @submit="handleSubmit"
    />

    <PaymentMethodDeleteDialog
      v-model:open="deleteOpen"
      :method="deleteTarget"
      :submitting="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>
