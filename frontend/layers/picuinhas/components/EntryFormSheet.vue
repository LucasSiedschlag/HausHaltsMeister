<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import { Switch } from '~/layers/shared/components/ui/switch'
import type { CardOwner, PaymentMethod, PicuinhaEntry, PicuinhaKind, Person } from '../types/picuinha'
import { validateEntryInput } from '../validation/entry'

interface Props {
  open: boolean
  mode: 'create' | 'edit'
  entry?: PicuinhaEntry | null
  persons: Person[]
  paymentMethods: PaymentMethod[]
  submitting?: boolean
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  entry: null,
  paymentMethods: () => [],
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { person_id: number; amount: number; kind: PicuinhaKind; auto_create_flow: boolean }]
}>()

const form = reactive<{
  person_id: string
  amount: string
  kind: PicuinhaKind
  auto_create_flow: boolean
  payment_method_id: string
  card_owner: CardOwner
}>({
  person_id: '',
  amount: '',
  kind: 'PLUS',
  auto_create_flow: true,
  payment_method_id: '',
  card_owner: 'SELF',
})

const errors = ref<{ person_id?: string; amount?: string; kind?: string }>({})

const sheetTitle = computed(() => (props.mode === 'edit' ? 'Editar lançamento' : 'Novo lançamento'))

function resetForm() {
  if (props.entry && props.mode === 'edit') {
    form.person_id = String(props.entry.person_id)
    form.amount = String(props.entry.amount)
    form.kind = props.entry.kind
    form.auto_create_flow = Boolean(props.entry.cash_flow_id)
    form.payment_method_id = props.entry.payment_method_id ? String(props.entry.payment_method_id) : ''
    form.card_owner = props.entry.card_owner || 'SELF'
  } else {
    form.person_id = ''
    form.amount = ''
    form.kind = 'PLUS'
    form.auto_create_flow = true
    form.payment_method_id = ''
    form.card_owner = 'SELF'
  }
  errors.value = {}
}

function handleSubmit() {
  const parsedPersonId = Number(form.person_id)
  const parsedAmount = Number(form.amount)
  const parsedPaymentMethod = form.payment_method_id ? Number(form.payment_method_id) : NaN
  const payload = {
    person_id: Number.isNaN(parsedPersonId) ? 0 : parsedPersonId,
    amount: Number.isNaN(parsedAmount) ? 0 : parsedAmount,
    kind: form.kind,
    auto_create_flow: form.auto_create_flow,
    payment_method_id: Number.isNaN(parsedPaymentMethod) ? undefined : parsedPaymentMethod,
    card_owner: form.card_owner,
  }
  const result = validateEntryInput(payload)
  errors.value = result.errors
  if (!result.valid) return
  emit('submit', payload)
}

watch(
  () => [props.open, props.entry, props.mode],
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
          <Label for="entry-person">Pessoa</Label>
          <Select id="entry-person" v-model="form.person_id" :disabled="props.submitting">
            <option value="" disabled>Selecione...</option>
            <option v-for="person in props.persons" :key="person.id" :value="String(person.id)">
              {{ person.name }}
            </option>
          </Select>
          <p v-if="errors.person_id" class="text-xs text-destructive">{{ errors.person_id }}</p>
        </div>

        <div class="space-y-2">
          <Label for="entry-kind">Tipo</Label>
          <Select id="entry-kind" v-model="form.kind" :disabled="props.submitting">
            <option value="PLUS">Emprestei (ela me deve)</option>
            <option value="MINUS">Recebi (ela pagou)</option>
          </Select>
          <p v-if="errors.kind" class="text-xs text-destructive">{{ errors.kind }}</p>
        </div>

        <div class="space-y-2">
          <Label for="entry-amount">Valor</Label>
          <Input
            id="entry-amount"
            v-model="form.amount"
            type="number"
            min="0"
            step="0.01"
            :disabled="props.submitting"
            placeholder="Ex: 120.00"
          />
          <p v-if="errors.amount" class="text-xs text-destructive">{{ errors.amount }}</p>
        </div>

        <div class="space-y-2">
          <Label for="entry-card">Cartão (meu)</Label>
          <Select id="entry-card" v-model="form.payment_method_id" :disabled="props.submitting">
            <option value="">Sem cartão</option>
            <option v-for="method in props.paymentMethods" :key="method.id" :value="String(method.id)">
              {{ method.name }}
            </option>
          </Select>
          <p class="text-xs text-muted-foreground">Opcional: registra o cartão utilizado.</p>
        </div>

        <div class="flex items-center justify-between rounded-md border p-3">
          <div>
            <p class="text-sm font-medium">Criar fluxo automático</p>
            <p class="text-xs text-muted-foreground">Gera um lançamento no fluxo de caixa.</p>
          </div>
          <Switch v-model="form.auto_create_flow" :disabled="props.submitting" />
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
