<script setup lang="ts">
import { Button } from '@shared/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { useMonthPeriod } from '@shared/composables/useMonthPeriod'
import { computed } from 'vue'

const props = defineProps<{
  contextId?: string
}>()

const { locale } = useI18n()
const { year, month, setMonth, setYear, nextMonth, previousMonth } = useMonthPeriod(props.contextId)

const months = computed(() => {
  const formatter = new Intl.DateTimeFormat(locale.value, { month: 'long' })
  return Array.from({ length: 12 }, (_, i) => ({
    value: i + 1,
    label: formatter.format(new Date(2000, i, 1))
  }))
})

const yearOptions = computed(() => {
  const current = new Date().getFullYear()
  const start = current - 5
  return Array.from({ length: 10 }, (_, i) => start + i)
})
</script>

<template>
  <div class="flex items-center gap-2">
    <Button variant="outline" size="icon" class="h-9 w-9" @click="previousMonth">
      <ChevronLeft class="h-4 w-4" />
    </Button>

    <div class="flex items-center gap-1">
      <Select :model-value="String(month)" @update:model-value="(v) => setMonth(Number(v))">
        <SelectTrigger class="h-9 w-[130px] capitalize">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="m in months" :key="m.value" :value="String(m.value)" class="capitalize">
            {{ m.label }}
          </SelectItem>
        </SelectContent>
      </Select>

      <Select :model-value="String(year)" @update:model-value="(v) => setYear(Number(v))">
        <SelectTrigger class="h-9 w-[90px]">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="y in yearOptions" :key="y" :value="String(y)">
            {{ y }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <Button variant="outline" size="icon" class="h-9 w-9" @click="nextMonth">
      <ChevronRight class="h-4 w-4" />
    </Button>
  </div>
</template>
