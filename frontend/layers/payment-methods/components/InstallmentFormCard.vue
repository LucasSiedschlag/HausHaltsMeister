<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import type { InstallmentCategoryOption } from '../types/installment'
import type { PaymentMethod } from '../types/payment-method'
import type { Person } from '~/layers/picuinhas/types/picuinha'
import { validateInstallmentInput } from '../validation/installment'

interface Props {
  categories: InstallmentCategoryOption[]
  paymentMethods: PaymentMethod[]
  persons?: Person[]
  categoriesLoading?: boolean
  paymentMethodsLoading?: boolean
  submitting?: boolean
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  categories: () => [],
  paymentMethods: () => [],
  persons: () => [],
  categoriesLoading: false,
  paymentMethodsLoading: false,
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  submit: [payload: {
    description: string
    amount_mode: 'TOTAL' | 'INSTALLMENT'
    total_amount?: number
    installment_amount?: number
    count: number
    category_id: number
    payment_method_id: number
    purchase_date: string
    person_id?: number
  }]
}>()

const form = reactive({
  description: '',
  amount_mode: 'TOTAL' as 'TOTAL' | 'INSTALLMENT',
  total_amount: '',
  installment_amount: '',
  count: '',
  category_id: '',
  payment_method_id: '',
  purchase_date: '',
  person_id: '',
})

const errors = ref<Record<string, string | undefined>>({})

const cardMethods = computed(() =>
  props.paymentMethods.filter((method) => method.kind === 'CREDIT_CARD' && method.is_active),
)

const outCategories = computed(() =>
  props.categories.filter((category) => category.direction === 'OUT' && category.is_active),
)

function setAmountMode(mode: 'TOTAL' | 'INSTALLMENT') {
  form.amount_mode = mode
}

function handleSubmit() {
  const result = validateInstallmentInput({
    description: form.description,
    amount_mode: form.amount_mode,
    total_amount: form.total_amount,
    installment_amount: form.installment_amount,
    count: form.count,
    category_id: form.category_id,
    payment_method_id: form.payment_method_id,
    purchase_date: form.purchase_date,
  })
  errors.value = result.errors
  if (!result.valid) return
  emit('submit', {
    ...result.values,
    person_id: form.person_id ? Number(form.person_id) : undefined,
  })
}

watch(
  () => [props.categoriesLoading, props.paymentMethodsLoading],
  () => {
    if (props.categoriesLoading || props.paymentMethodsLoading) {
      errors.value = {}
    }
  },
)
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
    <div class="border-b px-6 py-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="text-base font-semibold">Nova compra parcelada</h2>
          <p class="text-sm text-muted-foreground">Registre compras no cartão e gere parcelas automáticas.</p>
        </div>
        <Button type="button" :disabled="props.submitting" @click="handleSubmit">
          Registrar parcelamento
        </Button>
      </div>
    </div>

    <div class="space-y-4 px-6 py-6">
      <div v-if="props.errorMessage" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
        {{ props.errorMessage }}
      </div>

      <div class="space-y-2">
        <Label for="installment-description">Descrição</Label>
        <Input
          id="installment-description"
          v-model="form.description"
          :disabled="props.submitting"
          placeholder="Ex: TV Samsung 55"
        />
        <p v-if="errors.description" class="text-xs text-destructive">{{ errors.description }}</p>
      </div>

      <div class="space-y-2">
        <Label for="installment-person">Pessoa (picuinha)</Label>
        <Select id="installment-person" v-model="form.person_id" :disabled="props.submitting">
          <option value="">Sem pessoa selecionada</option>
          <option v-for="person in props.persons" :key="person.id" :value="String(person.id)">
            {{ person.name }}
          </option>
        </Select>
        <p class="text-xs text-muted-foreground">Selecione se a compra pertence a alguém.</p>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <Label for="installment-card">Cartão</Label>
          <Select id="installment-card" v-model="form.payment_method_id" :disabled="props.submitting || props.paymentMethodsLoading">
            <option value="">Selecione um cartão</option>
            <option v-for="method in cardMethods" :key="method.id" :value="String(method.id)">
              {{ method.name }}
            </option>
          </Select>
          <p v-if="errors.payment_method_id" class="text-xs text-destructive">{{ errors.payment_method_id }}</p>
          <p v-if="!cardMethods.length" class="text-xs text-muted-foreground">
            Nenhum cartão de crédito disponível.
          </p>
        </div>

        <div class="space-y-2">
          <Label for="installment-category">Categoria</Label>
          <Select id="installment-category" v-model="form.category_id" :disabled="props.submitting || props.categoriesLoading">
            <option value="">Selecione uma categoria</option>
            <option v-for="category in outCategories" :key="category.id" :value="String(category.id)">
              {{ category.name }}
            </option>
          </Select>
          <p v-if="errors.category_id" class="text-xs text-destructive">{{ errors.category_id }}</p>
          <p v-if="!outCategories.length" class="text-xs text-muted-foreground">
            Nenhuma categoria de saída ativa disponível.
          </p>
        </div>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <Label for="installment-date">Data da compra</Label>
          <Input
            id="installment-date"
            v-model="form.purchase_date"
            type="date"
            :disabled="props.submitting"
          />
          <p v-if="errors.purchase_date" class="text-xs text-destructive">{{ errors.purchase_date }}</p>
        </div>

        <div class="space-y-2">
          <Label for="installment-count">Parcelas</Label>
          <Input
            id="installment-count"
            v-model="form.count"
            type="number"
            inputmode="numeric"
            min="1"
            step="1"
            :disabled="props.submitting"
            placeholder="Ex: 10"
          />
          <p v-if="errors.count" class="text-xs text-destructive">{{ errors.count }}</p>
        </div>
      </div>

      <div class="flex flex-col gap-2">
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
        <p v-if="errors.amount_mode" class="text-xs text-destructive">{{ errors.amount_mode }}</p>
      </div>

      <div v-if="form.amount_mode === 'TOTAL'" class="space-y-2">
        <Label for="installment-total">Valor total</Label>
        <Input
          id="installment-total"
          v-model="form.total_amount"
          type="number"
          inputmode="decimal"
          min="0"
          step="0.01"
          :disabled="props.submitting"
          placeholder="Ex: 2500,00"
        />
        <p v-if="errors.total_amount" class="text-xs text-destructive">{{ errors.total_amount }}</p>
      </div>

      <div v-else class="space-y-2">
        <Label for="installment-amount">Valor da parcela</Label>
        <Input
          id="installment-amount"
          v-model="form.installment_amount"
          type="number"
          inputmode="decimal"
          min="0"
          step="0.01"
          :disabled="props.submitting"
          placeholder="Ex: 299,90"
        />
        <p v-if="errors.installment_amount" class="text-xs text-destructive">{{ errors.installment_amount }}</p>
      </div>

      <div class="flex justify-end gap-2">
        <Button :disabled="props.submitting" @click="handleSubmit">
          {{ props.submitting ? 'Salvando...' : 'Registrar parcelamento' }}
        </Button>
      </div>
    </div>
  </div>
</template>
