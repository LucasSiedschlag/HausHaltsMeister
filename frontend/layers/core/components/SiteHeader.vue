<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@shared/components/ui/button'
import ButtonGroup from '@shared/components/ui/button-group/ButtonGroup.vue'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Separator } from '@shared/components/ui/separator'
import { SidebarTrigger } from '@shared/components/ui/sidebar'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { useJournalUi } from '#layers/journal/composables/useJournalUi'
import { useJournalPeriod } from '#layers/journal/composables/useJournalPeriod'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import LedgerSwitcher from './LedgerSwitcher.vue'
import { cn } from '@shared/utils'

const { locale, t } = useI18n()
const ledgerContext = useLedgerContext()
const journalUi = useJournalUi()
const journalPeriod = useJournalPeriod()
const route = useRoute()
const router = useRouter()

const canEdit = computed(() => ledgerContext.hasRole('editor'))
const showJournalPeriod = computed(() => route.path.startsWith('/journal'))
const yearOptions = computed(() => {
  const current = journalPeriod.year.value
  return [current - 2, current - 1, current, current + 1, current + 2]
})
const selectedYear = computed({
  get: () => String(journalPeriod.year.value),
  set: (value: string) => journalPeriod.setYear(Number(value))
})
const monthLabels = computed(() =>
  Array.from({ length: 12 }, (_, index) => {
    const label = new Intl.DateTimeFormat(locale.value, { month: 'short' }).format(new Date(2026, index, 1))
    if (locale.value.startsWith('pt')) {
      return label.replace('.', '')
    }
    return label
  })
)

const monthButtonClass = (index: number) => {
  const isActive = journalPeriod.month.value === index + 1
  return cn(
    'h-8 px-3 text-[11px] uppercase transition-colors focus-visible:z-10 rounded-none first:rounded-l-md last:rounded-r-md border border-border/60',
    isActive
      ? 'bg-primary text-primary-foreground border-primary hover:bg-primary/90 dark:hover:bg-primary/70'
      : 'bg-background text-muted-foreground hover:bg-muted/30'
  )
}

const goPrevMonth = () => {
  if (journalPeriod.month.value === 1) {
    journalPeriod.setMonth(12)
    journalPeriod.setYear(journalPeriod.year.value - 1)
    return
  }
  journalPeriod.setMonth(journalPeriod.month.value - 1)
}

const goNextMonth = () => {
  if (journalPeriod.month.value === 12) {
    journalPeriod.setMonth(1)
    journalPeriod.setYear(journalPeriod.year.value + 1)
    return
  }
  journalPeriod.setMonth(journalPeriod.month.value + 1)
}

const handleNewTransaction = async () => {
  journalUi.openCreate()
  if (route.path !== '/journal') {
    await router.push('/journal')
  }
}
</script>

<template>
  <header class="sticky top-0 z-30 flex h-16 items-center justify-center gap-4 border-b border-border/60 bg-background/80 px-6 backdrop-blur">
    <SidebarTrigger class="lg:hidden" />
    <Separator orientation="vertical" class="h-6 lg:hidden" />
    <LedgerSwitcher />
    <div v-if="showJournalPeriod" class="flex flex-wrap items-center gap-2">
      <Select v-model="selectedYear">
        <SelectTrigger class="h-9 w-[110px]">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="year in yearOptions" :key="year" :value="String(year)">
            {{ year }}
          </SelectItem>
        </SelectContent>
      </Select>
      <div class="flex items-center gap-1">
        <Button
          type="button"
          size="icon-sm"
          variant="outline"
          class="h-8 w-8 border"
          @click="goPrevMonth"
        >
          <ChevronLeft class="h-4 w-4" />
        </Button>
        <ButtonGroup class="h-8">
          <Button
            v-for="(label, index) in monthLabels"
            :key="label"
            type="button"
            size="sm"
            variant="ghost"
            class="first:rounded-l-md last:rounded-r-md"
            :class="monthButtonClass(index)"
            @click="journalPeriod.setMonth(index + 1)"
          >
            {{ label }}
          </Button>
        </ButtonGroup>
        <Button
          type="button"
          size="icon-sm"
          variant="outline"
          class="h-8 w-8 border"
          @click="goNextMonth"
        >
          <ChevronRight class="h-4 w-4" />
        </Button>
      </div>
    </div>
    <div class="flex items-center gap-2">
        <Button :disabled="!canEdit" @click="handleNewTransaction">
          {{ t('journal.header.newTransaction') }}
        </Button>
    </div>
  </header>
</template>
