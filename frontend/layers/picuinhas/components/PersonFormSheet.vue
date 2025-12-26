<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from '~/layers/shared/components/ui/sheet'
import { Button } from '~/layers/shared/components/ui/button'
import { Input } from '~/layers/shared/components/ui/input'
import { Label } from '~/layers/shared/components/ui/label'
import type { Person } from '../types/picuinha'
import { validatePersonInput } from '../validation/person'

interface Props {
  open: boolean
  mode: 'create' | 'edit'
  person?: Person | null
  submitting?: boolean
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  person: null,
  submitting: false,
  errorMessage: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { name: string; notes: string }]
}>()

const form = reactive<{ name: string; notes: string }>({
  name: '',
  notes: '',
})

const errors = ref<{ name?: string }>({})

const sheetTitle = computed(() => (props.mode === 'edit' ? 'Editar pessoa' : 'Nova pessoa'))

function resetForm() {
  if (props.person && props.mode === 'edit') {
    form.name = props.person.name
    form.notes = props.person.notes || ''
  } else {
    form.name = ''
    form.notes = ''
  }
  errors.value = {}
}

function handleSubmit() {
  const result = validatePersonInput({
    name: form.name,
    notes: form.notes,
  })
  errors.value = result.errors
  if (!result.valid) return
  emit('submit', {
    name: form.name.trim(),
    notes: form.notes.trim(),
  })
}

watch(
  () => [props.open, props.person, props.mode],
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
          <Label for="person-name">Nome</Label>
          <Input
            id="person-name"
            v-model="form.name"
            :disabled="props.submitting"
            placeholder="Ex: João Silva"
          />
          <p v-if="errors.name" class="text-xs text-destructive">{{ errors.name }}</p>
        </div>

        <div class="space-y-2">
          <Label for="person-notes">Observações</Label>
          <Input
            id="person-notes"
            v-model="form.notes"
            :disabled="props.submitting"
            placeholder="Ex: Amigo do trabalho"
          />
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
