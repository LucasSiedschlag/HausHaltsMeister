<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import { Switch } from '~/layers/shared/components/ui/switch'
import type { PaymentMethod, PaymentMethodKind } from '../types/payment-method'
import { validatePaymentMethodInput } from '../validation/payment-method'

interface Props {
  open: boolean
  mode: 'create' | 'edit'
  method?: PaymentMethod | null
  submitting?: boolean
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  method: null,
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { name: string; kind: PaymentMethodKind; bank_name: string; closing_day?: number | null; due_day?: number | null; is_active?: boolean }]
}>()

const form = reactive<{
  name: string
  kind: PaymentMethodKind
  bank_name: string
  closing_day: string
  due_day: string
  is_active: boolean
}>({
  name: '',
  kind: 'CREDIT_CARD',
  bank_name: '',
  closing_day: '',
  due_day: '',
  is_active: true,
})

const errors = ref<{ name?: string; kind?: string; closing_day?: string; due_day?: string }>({})

const sheetTitle = computed(() => (props.mode === 'edit' ? 'Editar meio de pagamento' : 'Novo meio de pagamento'))
const isCreditCard = computed(() => form.kind === 'CREDIT_CARD')
const initialActive = ref(true)

function resetForm() {
  if (props.method && props.mode === 'edit') {
    form.name = props.method.name
    form.kind = props.method.kind as PaymentMethodKind
    form.bank_name = props.method.bank_name || ''
    form.closing_day = props.method.closing_day ? String(props.method.closing_day) : ''
    form.due_day = props.method.due_day ? String(props.method.due_day) : ''
    form.is_active = props.method.is_active
    initialActive.value = props.method.is_active
  } else {
    form.name = ''
    form.kind = 'CREDIT_CARD'
    form.bank_name = ''
    form.closing_day = ''
    form.due_day = ''
    form.is_active = true
    initialActive.value = true
  }
  errors.value = {}
}

function handleSubmit() {
  const result = validatePaymentMethodInput(form)
  errors.value = result.errors
  if (!result.valid) return
  emit('submit', {
    ...result.values,
    is_active: props.mode === 'edit' ? form.is_active : true,
  })
}

watch(
  () => [props.open, props.method, props.mode],
  ([open]) => {
    if (open) resetForm()
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
          <Label for="payment-method-name">Nome</Label>
          <Input
            id="payment-method-name"
            v-model="form.name"
            :disabled="props.submitting"
            placeholder="Ex: Nubank, Carteira"
          />
          <p v-if="errors.name" class="text-xs text-destructive">{{ errors.name }}</p>
        </div>

        <div class="space-y-2">
          <Label for="payment-method-kind">Tipo</Label>
          <Select id="payment-method-kind" v-model="form.kind" :disabled="props.submitting">
            <option value="CREDIT_CARD">Cartão de crédito</option>
            <option value="DEBIT_CARD">Cartão de débito</option>
            <option value="PIX">PIX</option>
            <option value="CASH">Dinheiro</option>
            <option value="BANK_SLIP">Boleto</option>
          </Select>
          <p v-if="errors.kind" class="text-xs text-destructive">{{ errors.kind }}</p>
        </div>

        <div class="space-y-2">
          <Label for="payment-method-bank">Banco/Emissor</Label>
          <Input
            id="payment-method-bank"
            v-model="form.bank_name"
            :disabled="props.submitting"
            placeholder="Ex: Nubank, Itaú"
          />
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="payment-method-closing">Dia de fechamento</Label>
            <Input
              id="payment-method-closing"
              v-model="form.closing_day"
              type="number"
              inputmode="numeric"
              min="1"
              max="31"
              step="1"
              :disabled="props.submitting || !isCreditCard"
              placeholder="Ex: 1"
            />
            <p v-if="errors.closing_day" class="text-xs text-destructive">{{ errors.closing_day }}</p>
            <p v-else-if="!isCreditCard" class="text-xs text-muted-foreground">
              Disponível apenas para cartão de crédito.
            </p>
          </div>

          <div class="space-y-2">
            <Label for="payment-method-due">Dia de vencimento</Label>
            <Input
              id="payment-method-due"
              v-model="form.due_day"
              type="number"
              inputmode="numeric"
              min="1"
              max="31"
              step="1"
              :disabled="props.submitting || !isCreditCard"
              placeholder="Ex: 7"
            />
            <p v-if="errors.due_day" class="text-xs text-destructive">{{ errors.due_day }}</p>
            <p v-else-if="!isCreditCard" class="text-xs text-muted-foreground">
              Disponível apenas para cartão de crédito.
            </p>
          </div>
        </div>

        <div v-if="props.mode === 'edit'" class="rounded-md border p-3">
          <div class="flex items-center justify-between gap-4">
            <div>
              <p class="text-sm font-medium">
                {{
                  initialActive
                    ? 'Meio ativo'
                    : form.is_active
                      ? 'Reativar meio de pagamento'
                      : 'Meio inativo'
                }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{
                  initialActive
                    ? 'Para desativar, use o botão Desativar na lista.'
                    : form.is_active
                      ? 'O meio será reativado quando você salvar.'
                      : 'Reative para voltar a usar nos lançamentos.'
                }}
              </p>
            </div>
            <Switch
              v-if="!initialActive"
              v-model="form.is_active"
              :disabled="props.submitting"
              aria-label="Reativar meio de pagamento"
            />
          </div>
        </div>
      </div>

      <SheetFooter class="mt-6">
        <Button variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button :disabled="props.submitting" @click="handleSubmit">
          {{ props.mode === 'edit' ? 'Salvar' : 'Criar' }}
        </Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
