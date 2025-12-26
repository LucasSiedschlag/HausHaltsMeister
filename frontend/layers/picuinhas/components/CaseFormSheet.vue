<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import type { PicuinhaCaseType, PaymentMethod } from '../types/picuinha'
import type { Category } from '~/layers/categories/types/category'

interface FormPayload {
  title: string
  case_type: PicuinhaCaseType
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
}

interface Props {
  open: boolean
  submitting?: boolean
  errorMessage?: string | null
  paymentMethods: PaymentMethod[]
  categories: Category[]
}

const props = withDefaults(defineProps<Props>(), {
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: FormPayload]
}>()

const form = reactive({
  title: '',
  case_type: 'ONE_OFF' as PicuinhaCaseType,
  amount_mode: 'TOTAL' as 'TOTAL' | 'INSTALLMENT',
  total_amount: '',
  installment_amount: '',
  installment_count: '',
  start_date: '',
  purchase_date: '',
  payment_method_id: '',
  category_id: '',
  interest_rate: '',
  interest_rate_unit: 'MONTHLY',
  recurrence_interval_months: '1',
})

const errors = ref<Record<string, string | undefined>>({})

const cardMethods = computed(() =>
  props.paymentMethods.filter((method) => method.kind === 'CREDIT_CARD' && method.is_active),
)

const outCategories = computed(() =>
  props.categories.filter((category) => category.direction === 'OUT' && category.is_active),
)

const sheetTitle = computed(() => 'Nova picuinha')

const isInstallment = computed(() => form.case_type === 'INSTALLMENT' || form.case_type === 'CARD_INSTALLMENT')
const isCardInstallment = computed(() => form.case_type === 'CARD_INSTALLMENT')
const isRecurring = computed(() => form.case_type === 'RECURRING')

function resetForm() {
  form.title = ''
  form.case_type = 'ONE_OFF'
  form.amount_mode = 'TOTAL'
  form.total_amount = ''
  form.installment_amount = ''
  form.installment_count = ''
  form.start_date = ''
  form.purchase_date = ''
  form.payment_method_id = ''
  form.category_id = ''
  form.interest_rate = ''
  form.interest_rate_unit = 'MONTHLY'
  form.recurrence_interval_months = '1'
  errors.value = {}
}

function setAmountMode(mode: 'TOTAL' | 'INSTALLMENT') {
  form.amount_mode = mode
}

function handleSubmit() {
  const nextErrors: Record<string, string> = {}

  if (!form.title.trim()) {
    nextErrors.title = 'Informe um título.'
  }

  if (!form.case_type) {
    nextErrors.case_type = 'Selecione o tipo.'
  }

  if (form.case_type === 'ONE_OFF') {
    if (!Number(form.total_amount) || Number(form.total_amount) <= 0) {
      nextErrors.total_amount = 'Informe um valor válido.'
    }
    if (!form.start_date) {
      nextErrors.start_date = 'Informe a data.'
    }
  }

  if (isInstallment.value) {
    if (!Number(form.installment_count) || Number(form.installment_count) <= 0) {
      nextErrors.installment_count = 'Informe o número de parcelas.'
    }
    if (form.amount_mode === 'TOTAL' && (!Number(form.total_amount) || Number(form.total_amount) <= 0)) {
      nextErrors.total_amount = 'Informe o valor total.'
    }
    if (form.amount_mode === 'INSTALLMENT' && (!Number(form.installment_amount) || Number(form.installment_amount) <= 0)) {
      nextErrors.installment_amount = 'Informe o valor da parcela.'
    }
  }

  if (isInstallment.value && !isCardInstallment.value && !form.start_date) {
    nextErrors.start_date = 'Informe a data da primeira parcela.'
  }

  if (isCardInstallment.value) {
    if (!form.payment_method_id) {
      nextErrors.payment_method_id = 'Selecione um cartão.'
    }
    if (!form.category_id) {
      nextErrors.category_id = 'Selecione uma categoria.'
    }
    if (!form.purchase_date) {
      nextErrors.purchase_date = 'Informe a data da compra.'
    }
  }

  if (isRecurring.value) {
    if (!Number(form.installment_amount) || Number(form.installment_amount) <= 0) {
      nextErrors.installment_amount = 'Informe o valor mensal.'
    }
    if (!form.start_date) {
      nextErrors.start_date = 'Informe a data inicial.'
    }
    if (!Number(form.recurrence_interval_months) || Number(form.recurrence_interval_months) <= 0) {
      nextErrors.recurrence_interval_months = 'Intervalo inválido.'
    }
  }

  if (form.interest_rate && Number(form.interest_rate) > 0 && !form.interest_rate_unit) {
    nextErrors.interest_rate_unit = 'Selecione o tipo de taxa.'
  }

  errors.value = nextErrors
  if (Object.keys(nextErrors).length) {
    return
  }

  const payload: FormPayload = {
    title: form.title.trim(),
    case_type: form.case_type,
    amount_mode: form.amount_mode,
    total_amount: form.total_amount ? Number(form.total_amount) : undefined,
    installment_amount: form.installment_amount ? Number(form.installment_amount) : undefined,
    installment_count: form.installment_count ? Number(form.installment_count) : undefined,
    start_date: form.start_date || undefined,
    purchase_date: form.purchase_date || undefined,
    payment_method_id: form.payment_method_id ? Number(form.payment_method_id) : undefined,
    category_id: form.category_id ? Number(form.category_id) : undefined,
    interest_rate: form.interest_rate ? Number(form.interest_rate) : undefined,
    interest_rate_unit: form.interest_rate ? form.interest_rate_unit : undefined,
    recurrence_interval_months: form.recurrence_interval_months ? Number(form.recurrence_interval_months) : undefined,
  }

  emit('submit', payload)
}

watch(
  () => props.open,
  (open) => {
    if (open) resetForm()
  },
)
</script>

<template>
  <Sheet :open="props.open" @update:open="emit('update:open', $event)">
    <SheetContent side="right" class="w-full sm:max-w-lg">
      <SheetHeader>
        <SheetTitle>{{ sheetTitle }}</SheetTitle>
      </SheetHeader>

      <div class="mt-6 space-y-4">
        <div v-if="props.errorMessage" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
          {{ props.errorMessage }}
        </div>

        <div class="space-y-2">
          <Label for="case-title">Título</Label>
          <Input id="case-title" v-model="form.title" :disabled="props.submitting" placeholder="Ex: TV Samsung, Pix emergencial" />
          <p v-if="errors.title" class="text-xs text-destructive">{{ errors.title }}</p>
        </div>

        <div class="space-y-2">
          <Label for="case-type">Tipo</Label>
          <Select id="case-type" v-model="form.case_type" :disabled="props.submitting">
            <option value="ONE_OFF">Única</option>
            <option value="INSTALLMENT">Parcelada</option>
            <option value="CARD_INSTALLMENT">Compra no cartão</option>
            <option value="RECURRING">Recorrente</option>
          </Select>
          <p v-if="errors.case_type" class="text-xs text-destructive">{{ errors.case_type }}</p>
        </div>

        <div v-if="isInstallment" class="flex flex-col gap-2">
          <Label>Modo de valor</Label>
          <div class="flex w-fit flex-wrap gap-1 rounded-md border bg-muted/30 p-1">
            <Button
              type="button"
              size="sm"
              :variant="form.amount_mode === 'TOTAL' ? 'secondary' : 'ghost'"
              class="rounded-sm"
              :disabled="props.submitting"
              @click="setAmountMode('TOTAL')"
            >
              Valor total
            </Button>
            <Button
              type="button"
              size="sm"
              :variant="form.amount_mode === 'INSTALLMENT' ? 'secondary' : 'ghost'"
              class="rounded-sm"
              :disabled="props.submitting"
              @click="setAmountMode('INSTALLMENT')"
            >
              Valor da parcela
            </Button>
          </div>
        </div>

        <div v-if="form.case_type === 'ONE_OFF'" class="space-y-2">
          <Label for="case-total">Valor</Label>
          <Input id="case-total" v-model="form.total_amount" type="number" min="0" step="0.01" :disabled="props.submitting" />
          <p v-if="errors.total_amount" class="text-xs text-destructive">{{ errors.total_amount }}</p>
        </div>

        <div v-if="isInstallment && form.amount_mode === 'TOTAL'" class="space-y-2">
          <Label for="case-total-installment">Valor total</Label>
          <Input id="case-total-installment" v-model="form.total_amount" type="number" min="0" step="0.01" :disabled="props.submitting" />
          <p v-if="errors.total_amount" class="text-xs text-destructive">{{ errors.total_amount }}</p>
        </div>

        <div v-if="isInstallment && form.amount_mode === 'INSTALLMENT'" class="space-y-2">
          <Label for="case-installment-amount">Valor da parcela</Label>
          <Input id="case-installment-amount" v-model="form.installment_amount" type="number" min="0" step="0.01" :disabled="props.submitting" />
          <p v-if="errors.installment_amount" class="text-xs text-destructive">{{ errors.installment_amount }}</p>
        </div>

        <div v-if="isInstallment" class="space-y-2">
          <Label for="case-installment-count">Parcelas</Label>
          <Input id="case-installment-count" v-model="form.installment_count" type="number" min="1" step="1" :disabled="props.submitting" />
          <p v-if="errors.installment_count" class="text-xs text-destructive">{{ errors.installment_count }}</p>
        </div>

        <div v-if="isCardInstallment" class="grid gap-4 md:grid-cols-2">
          <div class="space-y-2">
            <Label for="case-card">Cartão</Label>
            <Select id="case-card" v-model="form.payment_method_id" :disabled="props.submitting">
              <option value="">Selecione um cartão</option>
              <option v-for="method in cardMethods" :key="method.id" :value="String(method.id)">
                {{ method.name }}
              </option>
            </Select>
            <p v-if="errors.payment_method_id" class="text-xs text-destructive">{{ errors.payment_method_id }}</p>
          </div>
          <div class="space-y-2">
            <Label for="case-category">Categoria</Label>
            <Select id="case-category" v-model="form.category_id" :disabled="props.submitting">
              <option value="">Selecione uma categoria</option>
              <option v-for="category in outCategories" :key="category.id" :value="String(category.id)">
                {{ category.name }}
              </option>
            </Select>
            <p v-if="errors.category_id" class="text-xs text-destructive">{{ errors.category_id }}</p>
          </div>
        </div>

        <div v-if="isCardInstallment" class="space-y-2">
          <Label for="case-purchase-date">Data da compra</Label>
          <Input id="case-purchase-date" v-model="form.purchase_date" type="date" :disabled="props.submitting" />
          <p v-if="errors.purchase_date" class="text-xs text-destructive">{{ errors.purchase_date }}</p>
        </div>

        <div v-if="isRecurring" class="space-y-2">
          <Label for="case-recurring-amount">Valor mensal</Label>
          <Input id="case-recurring-amount" v-model="form.installment_amount" type="number" min="0" step="0.01" :disabled="props.submitting" />
          <p v-if="errors.installment_amount" class="text-xs text-destructive">{{ errors.installment_amount }}</p>
        </div>

        <div v-if="!isCardInstallment" class="space-y-2">
          <Label for="case-start-date">Data inicial</Label>
          <Input id="case-start-date" v-model="form.start_date" type="date" :disabled="props.submitting" />
          <p v-if="errors.start_date" class="text-xs text-destructive">{{ errors.start_date }}</p>
        </div>

        <div v-if="isRecurring" class="space-y-2">
          <Label for="case-recurring-interval">Intervalo (meses)</Label>
          <Input id="case-recurring-interval" v-model="form.recurrence_interval_months" type="number" min="1" step="1" :disabled="props.submitting" />
          <p v-if="errors.recurrence_interval_months" class="text-xs text-destructive">{{ errors.recurrence_interval_months }}</p>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-2">
            <Label for="case-interest">Taxa de juros</Label>
            <Input id="case-interest" v-model="form.interest_rate" type="number" min="0" step="0.01" :disabled="props.submitting" />
          </div>
          <div class="space-y-2">
            <Label for="case-interest-unit">Tipo de taxa</Label>
            <Select id="case-interest-unit" v-model="form.interest_rate_unit" :disabled="props.submitting || !form.interest_rate">
              <option value="MONTHLY">Mensal</option>
              <option value="ANNUAL">Anual</option>
            </Select>
            <p v-if="errors.interest_rate_unit" class="text-xs text-destructive">{{ errors.interest_rate_unit }}</p>
          </div>
        </div>
      </div>

      <SheetFooter class="mt-6">
        <Button type="button" variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button type="button" :disabled="props.submitting" @click="handleSubmit">Salvar</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
