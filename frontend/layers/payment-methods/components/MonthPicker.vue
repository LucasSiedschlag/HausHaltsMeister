<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CalendarDays, ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { Button } from '~/layers/shared/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '~/layers/shared/components/ui/dropdown-menu'

interface Props {
  modelValue?: string
  id?: string
  disabled?: boolean
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  id: undefined,
  disabled: false,
  class: undefined,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const viewYear = ref(new Date().getFullYear())

const months = [
  { label: 'Jan', value: 1 },
  { label: 'Fev', value: 2 },
  { label: 'Mar', value: 3 },
  { label: 'Abr', value: 4 },
  { label: 'Mai', value: 5 },
  { label: 'Jun', value: 6 },
  { label: 'Jul', value: 7 },
  { label: 'Ago', value: 8 },
  { label: 'Set', value: 9 },
  { label: 'Out', value: 10 },
  { label: 'Nov', value: 11 },
  { label: 'Dez', value: 12 },
]

const selected = computed(() => {
  if (!props.modelValue) return null
  const [year, month] = props.modelValue.split('-')
  const parsedYear = Number(year)
  const parsedMonth = Number(month)
  if (!parsedYear || !parsedMonth) return null
  return { year: parsedYear, month: parsedMonth }
})

const displayLabel = computed(() => {
  if (!selected.value) return 'Selecionar mês'
  const date = new Date(selected.value.year, selected.value.month - 1, 1)
  const label = date.toLocaleDateString('pt-BR', { month: 'long', year: 'numeric' })
  return label.charAt(0).toUpperCase() + label.slice(1)
})

function selectMonth(month: number) {
  const year = viewYear.value
  const normalizedMonth = String(month).padStart(2, '0')
  emit('update:modelValue', `${year}-${normalizedMonth}`)
  open.value = false
}

function isSelected(month: number) {
  return selected.value?.year === viewYear.value && selected.value?.month === month
}

function previousYear() {
  viewYear.value -= 1
}

function nextYear() {
  viewYear.value += 1
}

watch(open, (value) => {
  if (value) {
    viewYear.value = selected.value?.year || new Date().getFullYear()
  }
})
</script>

<template>
  <DropdownMenu v-model:open="open">
    <DropdownMenuTrigger as-child>
      <Button
        :id="props.id"
        variant="outline"
        :class="['h-10 w-full justify-between gap-2 px-3', props.class]"
        :disabled="props.disabled"
      >
        <span class="truncate text-sm">{{ displayLabel }}</span>
        <CalendarDays class="h-4 w-4 text-muted-foreground" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" :side-offset="8" class="w-72 p-0">
      <div class="flex items-center justify-between border-b px-3 py-2">
        <Button variant="ghost" size="icon-sm" @click="previousYear">
          <ChevronLeft class="h-4 w-4" />
        </Button>
        <span class="text-sm font-semibold text-foreground">{{ viewYear }}</span>
        <Button variant="ghost" size="icon-sm" @click="nextYear">
          <ChevronRight class="h-4 w-4" />
        </Button>
      </div>
      <div class="grid grid-cols-3 gap-2 p-3">
        <button
          v-for="month in months"
          :key="month.value"
          type="button"
          class="rounded-md border px-2 py-2 text-sm font-medium transition-colors"
          :class="isSelected(month.value)
            ? 'border-primary/40 bg-primary/10 text-primary'
            : 'border-border bg-background text-foreground hover:bg-muted'"
          @click="selectMonth(month.value)"
        >
          {{ month.label }}
        </button>
      </div>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
