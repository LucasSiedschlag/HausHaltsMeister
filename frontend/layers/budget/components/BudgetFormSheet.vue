<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import type { BudgetItem, CategoryOption } from '../types/budget'
import { validateBudgetItemInput } from '../validation/budget'

interface Props {
  open: boolean
  mode: 'create' | 'edit'
  item?: BudgetItem | null
  categories: CategoryOption[]
  categoriesLoading?: boolean
  submitting?: boolean
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  item: null,
  categoriesLoading: false,
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { category_id: number; planned_amount: number }]
}>()

const form = reactive({
  category_id: '',
  planned_amount: '',
})

const errors = ref<{ category_id?: string; planned_amount?: string }>({})

const sheetTitle = computed(() => (props.mode === 'edit' ? 'Editar item' : 'Novo item'))

const categoryOptions = computed(() => {
  const options = [...props.categories]
  if (props.mode === 'edit' && props.item) {
    const exists = options.some((option) => option.id === props.item?.category_id)
    if (!exists) {
      options.unshift({
        id: props.item.category_id,
        name: props.item.category_name || `Categoria #${props.item.category_id}`,
        direction: 'OUT',
        is_active: false,
        is_budget_relevant: false,
      })
    }
  }
  return options
})

function resetForm() {
  if (props.mode === 'edit' && props.item) {
    form.category_id = String(props.item.category_id)
    form.planned_amount = props.item.planned_amount.toFixed(2)
  } else {
    form.category_id = ''
    form.planned_amount = ''
  }
  errors.value = {}
}

function handleSubmit() {
  const result = validateBudgetItemInput(form)
  errors.value = result.errors
  if (!result.valid) {
    return
  }
  emit('submit', result.values)
}

watch(
  () => [props.open, props.item, props.mode],
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
          <Label for="budget-category">Categoria</Label>
          <Select
            id="budget-category"
            v-model="form.category_id"
            :disabled="props.submitting || props.categoriesLoading || props.mode === 'edit'"
          >
            <option value="">Selecione uma categoria</option>
            <option v-for="category in categoryOptions" :key="category.id" :value="String(category.id)">
              {{ category.name }}
            </option>
          </Select>
          <p v-if="errors.category_id" class="text-xs text-destructive">{{ errors.category_id }}</p>
          <p v-if="!categoryOptions.length" class="text-xs text-muted-foreground">
            Nenhuma categoria elegível para orçamento.
          </p>
        </div>

        <div class="space-y-2">
          <Label for="budget-amount">Valor planejado</Label>
          <Input
            id="budget-amount"
            v-model="form.planned_amount"
            type="number"
            inputmode="decimal"
            min="0"
            step="0.01"
            :disabled="props.submitting"
            placeholder="Ex: 1500,00"
          />
          <p v-if="errors.planned_amount" class="text-xs text-destructive">{{ errors.planned_amount }}</p>
        </div>
      </div>

      <SheetFooter class="mt-6">
        <Button variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button :disabled="props.submitting || (props.mode === 'create' && !categoryOptions.length)" @click="handleSubmit">
          {{ props.mode === 'edit' ? 'Salvar' : 'Criar' }}
        </Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
