<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import { Select } from '~/layers/shared/components/ui/select'
import { Switch } from '~/layers/shared/components/ui/switch'
import type { Category, CreateCategoryRequest, Direction } from '../types/category'
import { validateCategoryInput } from '../validation/category'

interface Props {
  open: boolean
  mode: 'create' | 'edit'
  category?: Category | null
  submitting?: boolean
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  category: null,
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: CreateCategoryRequest]
}>()

const form = reactive<CreateCategoryRequest>({
  name: '',
  direction: 'OUT',
  is_budget_relevant: true,
})

const errors = ref<{ name?: string; direction?: string }>({})

const sheetTitle = computed(() => (props.mode === 'edit' ? 'Editar categoria' : 'Nova categoria'))

function resetForm() {
  if (props.category && props.mode === 'edit') {
    form.name = props.category.name
    form.direction = props.category.direction
    form.is_budget_relevant = props.category.is_budget_relevant
  } else {
    form.name = ''
    form.direction = 'OUT'
    form.is_budget_relevant = true
  }
  errors.value = {}
}

function handleSubmit() {
  const result = validateCategoryInput(form)
  errors.value = result.errors
  if (!result.valid) {
    return
  }
  emit('submit', {
    name: form.name.trim(),
    direction: form.direction as Direction,
    is_budget_relevant: form.is_budget_relevant,
  })
}

watch(
  () => [props.open, props.category, props.mode],
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
          <Label for="category-name">Nome</Label>
          <Input
            id="category-name"
            v-model="form.name"
            :disabled="props.submitting"
            placeholder="Ex: Alimentação"
          />
          <p v-if="errors.name" class="text-xs text-destructive">{{ errors.name }}</p>
        </div>

        <div class="space-y-2">
          <Label for="category-direction">Direção</Label>
          <Select
            id="category-direction"
            v-model="form.direction"
            :disabled="props.submitting"
          >
            <option value="IN">Entrada</option>
            <option value="OUT">Saída</option>
          </Select>
          <p v-if="errors.direction" class="text-xs text-destructive">{{ errors.direction }}</p>
        </div>

        <div class="flex items-center justify-between rounded-md border p-3">
          <div>
            <p class="text-sm font-medium">Relevante para orçamento</p>
            <p class="text-xs text-muted-foreground">Exibe a categoria no planejamento.</p>
          </div>
          <Switch v-model="form.is_budget_relevant" :disabled="props.submitting" />
        </div>

        <p v-if="props.mode === 'edit'" class="text-xs text-muted-foreground">
          Edição indisponível no backend no momento.
        </p>
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
