<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import type { CashflowEntry, CreateCashflowRequest, Direction } from '../types/cashflow'
import type { Category } from '~/layers/categories/types/category'
import type { PaymentMethod } from '~/layers/payment-methods/types/payment-method'
import { validateCashflowInput } from '../validation/cashflow'

interface Props {
  open: boolean
  mode: 'create' | 'edit'
  entry?: CashflowEntry | null
  categories: Category[]
  paymentMethods: PaymentMethod[]
  direction: Direction
  isFixed: boolean
  referenceDate: string
  categoryLocked?: boolean
  submitting?: boolean
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  entry: null,
  categoryLocked: false,
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: CreateCashflowRequest]
}>()

const form = reactive({
  title: '',
  amount: '',
  category_id: '',
  payment_method_id: '',
  date: '',
})

const errors = ref<Record<string, string | undefined>>({})

const sheetTitle = computed(() => (props.mode === 'edit' ? 'Editar lançamento' : 'Novo lançamento'))

const categoryOptions = computed(() => {
  const options = [...props.categories]
  if (props.mode === 'edit' && props.entry) {
    const exists = options.some((option) => option.id === props.entry?.category_id)
    if (!exists) {
      options.unshift({
        id: props.entry.category_id,
        name: props.entry.category_name || `Categoria #${props.entry.category_id}`,
        direction: props.direction,
        is_budget_relevant: false,
        is_active: false,
      })
    }
  }
  return options
})

const paymentOptions = computed(() => {
  const options = [...props.paymentMethods]
  if (props.mode === 'edit' && props.entry) {
    const exists = options.some((option) => option.id === props.entry?.payment_method_id)
    if (!exists) {
      options.unshift({
        id: props.entry.payment_method_id,
        name: props.entry.payment_method_name || `Meio #${props.entry.payment_method_id}`,
        kind: 'CASH',
        bank_name: '',
        is_active: false,
      })
    }
  }
  return options
})

const selectedPaymentMethod = computed(() => {
  if (!form.payment_method_id) return null
  return paymentOptions.value.find((method) => method.id === Number(form.payment_method_id)) || null
})

const requiresPurchaseDate = computed(() => selectedPaymentMethod.value?.kind === 'CREDIT_CARD')

const referenceLabel = computed(() => {
  if (!props.referenceDate) return ''
  const [year, month] = props.referenceDate.split('-').map(Number)
  if (!year || !month) return ''
  const label = new Date(year, month - 1, 1).toLocaleDateString('pt-BR', { month: 'long', year: 'numeric' })
  return label.charAt(0).toUpperCase() + label.slice(1)
})

watch(requiresPurchaseDate, (value) => {
  if (props.mode === 'edit' || props.entry) return
  if (value) {
    form.date = ''
  } else if (props.referenceDate) {
    form.date = props.referenceDate
  }
})

function resetForm() {
  if (props.mode === 'edit' && props.entry) {
    form.title = props.entry.title
    form.amount = props.entry.amount.toFixed(2)
    form.category_id = String(props.entry.category_id)
    form.payment_method_id = String(props.entry.payment_method_id)
    form.date = props.entry.date
  } else {
    form.title = ''
    form.amount = ''
    form.category_id = categoryOptions.value.length === 1 ? String(categoryOptions.value[0].id) : ''
    form.payment_method_id = ''
    form.date = props.referenceDate
  }
  errors.value = {}
}

function handleSubmit() {
  const result = validateCashflowInput(
    {
      title: form.title,
      amount: form.amount,
      category_id: form.category_id,
      payment_method_id: form.payment_method_id,
      date: form.date,
    },
    { requiresDate: requiresPurchaseDate.value },
  )
  errors.value = result.errors
  if (!result.valid) {
    return
  }

  const resolvedDate = requiresPurchaseDate.value ? result.values.date : props.referenceDate
  if (!resolvedDate) {
    errors.value = { ...errors.value, date: 'Selecione um mês válido.' }
    return
  }

  const payload: CreateCashflowRequest = {
    title: result.values.title,
    amount: result.values.amount,
    category_id: result.values.category_id,
    payment_method_id: result.values.payment_method_id,
    date: resolvedDate,
    direction: props.direction,
    is_fixed: props.isFixed,
  }

  emit('submit', payload)
}

watch(
  () => [props.open, props.entry, props.mode, props.referenceDate],
  ([open]) => {
    if (open) {
      resetForm()
    }
  },
)
</script>

<template>
  <Sheet :open="props.open" @update:open="emit('update:open', $event)">
    <SheetContent side="right" class="w-full sm:max-w-md">
      <SheetHeader>
        <SheetTitle>{{ sheetTitle }}</SheetTitle>
      </SheetHeader>

      <div class="mt-6 space-y-4">
        <div v-if="props.errorMessage" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
          {{ props.errorMessage }}
        </div>

        <div class="space-y-2">
          <Label for="cashflow-title">Título</Label>
          <Input
            id="cashflow-title"
            v-model="form.title"
            :disabled="props.submitting"
            placeholder="Ex: Salário, Mercado, Aluguel"
          />
          <p v-if="errors.title" class="text-xs text-destructive">{{ errors.title }}</p>
        </div>

        <div class="space-y-2">
          <Label for="cashflow-category">Categoria</Label>
          <Select
            id="cashflow-category"
            v-model="form.category_id"
            :disabled="props.submitting || props.categoryLocked"
          >
            <option value="">Selecione uma categoria</option>
            <option v-for="category in categoryOptions" :key="category.id" :value="String(category.id)">
              {{ category.name }}
            </option>
          </Select>
          <p v-if="errors.category_id" class="text-xs text-destructive">{{ errors.category_id }}</p>
          <p v-if="!categoryOptions.length" class="text-xs text-muted-foreground">
            Nenhuma categoria disponível para este tipo de lançamento.
          </p>
        </div>

        <div class="space-y-2">
          <Label for="cashflow-payment">Meio de pagamento</Label>
          <Select
            id="cashflow-payment"
            v-model="form.payment_method_id"
            :disabled="props.submitting"
          >
            <option value="">Selecione um meio</option>
            <option v-for="method in paymentOptions" :key="method.id" :value="String(method.id)">
              {{ method.name }}
            </option>
          </Select>
          <p v-if="errors.payment_method_id" class="text-xs text-destructive">{{ errors.payment_method_id }}</p>
          <p v-if="!paymentOptions.length" class="text-xs text-muted-foreground">
            Nenhum meio de pagamento ativo.
          </p>
        </div>

        <div class="space-y-2">
          <Label for="cashflow-amount">Valor</Label>
          <Input
            id="cashflow-amount"
            v-model="form.amount"
            type="number"
            inputmode="decimal"
            min="0"
            step="0.01"
            :disabled="props.submitting"
            placeholder="Ex: 120,00"
          />
          <p v-if="errors.amount" class="text-xs text-destructive">{{ errors.amount }}</p>
        </div>

        <div class="space-y-2">
          <Label v-if="requiresPurchaseDate" for="cashflow-date">Data da compra</Label>
          <Label v-else>Mês de referência</Label>
          <div v-if="requiresPurchaseDate" class="space-y-2">
            <Input
              id="cashflow-date"
              v-model="form.date"
              type="date"
              :disabled="props.submitting"
            />
            <p v-if="errors.date" class="text-xs text-destructive">{{ errors.date }}</p>
          </div>
          <p v-else class="text-sm text-muted-foreground">
            {{ referenceLabel || 'Selecione um mês válido.' }}
          </p>
        </div>
      </div>

      <SheetFooter class="mt-6">
        <Button variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button
          :disabled="props.submitting || (!categoryOptions.length && !props.entry)"
          @click="handleSubmit"
        >
          {{ props.mode === 'edit' ? 'Salvar' : 'Criar' }}
        </Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
