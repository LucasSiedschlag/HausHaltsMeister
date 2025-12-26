<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import InstallmentInvoiceCard from '../../components/InstallmentInvoiceCard.vue'
import type { InvoiceSummary } from '../../types/installment'
import type { PaymentMethod } from '../../types/payment-method'
import { useInstallmentsService } from '../../services/installments'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'

definePageMeta({
  layout: 'default',
})

const { listPaymentMethods, getInvoice } = useInstallmentsService()

const paymentMethods = ref<PaymentMethod[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)
const paymentMethodsLoading = ref(false)

const invoice = ref<InvoiceSummary | null>(null)
const invoiceLoading = ref(false)
const invoiceError = ref<string | null>(null)
const invoiceMonthValue = ref(getCurrentMonthValue())
const invoiceCardId = ref('')

const pageSubtitle = computed(() => 'Confira o total mensal e as parcelas do cartão.')

const invoiceMonthParam = computed(() =>
  invoiceMonthValue.value ? `${invoiceMonthValue.value}-01` : '',
)

function getCurrentMonthValue() {
  const date = new Date()
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  return `${year}-${month}`
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

async function refreshBaseData() {
  loading.value = true
  loadError.value = null
  await fetchPaymentMethods()
  loading.value = false
}

async function fetchInvoice() {
  if (!invoiceCardId.value) {
    invoiceError.value = 'Selecione um cartão para ver a fatura.'
    return
  }
  if (!invoiceMonthParam.value) {
    invoiceError.value = 'Selecione um mês de referência.'
    return
  }
  invoiceLoading.value = true
  invoiceError.value = null
  try {
    const id = Number(invoiceCardId.value)
    invoice.value = await getInvoice(id, invoiceMonthParam.value)
  } catch (error) {
    invoiceError.value = getApiErrorMessage(error)
  } finally {
    invoiceLoading.value = false
  }
}

watch([invoiceCardId, invoiceMonthValue], ([cardId, month]) => {
  invoice.value = null
  invoiceError.value = null
  if (cardId && month) {
    fetchInvoice()
  }
})

onMounted(refreshBaseData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Parcelamentos · Fatura de cartão</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" :disabled="loading" @click="refreshBaseData">Atualizar cartões</Button>
    </div>

    <div v-if="loadError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ loadError }}
    </div>

    <InstallmentInvoiceCard
      :payment-methods="paymentMethods"
      :payment-methods-loading="paymentMethodsLoading"
      :invoice="invoice"
      :loading="invoiceLoading"
      :error="invoiceError"
      :month-value="invoiceMonthValue"
      :card-id="invoiceCardId"
      @update:month-value="invoiceMonthValue = $event"
      @update:card-id="invoiceCardId = $event"
      @fetch="fetchInvoice"
    />
  </div>
</template>
