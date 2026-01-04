import { computed } from 'vue'

const toDateString = (date: Date) => date.toISOString().slice(0, 10)

export const useJournalPeriod = () => {
  const now = new Date()
  const year = useState<number>('journal_period_year', () => now.getFullYear())
  const month = useState<number>('journal_period_month', () => now.getMonth() + 1)

  const setYear = (value: number) => {
    year.value = value
  }

  const setMonth = (value: number) => {
    month.value = value
  }

  const range = computed(() => {
    const start = new Date(Date.UTC(year.value, month.value - 1, 1))
    const end = new Date(Date.UTC(year.value, month.value, 0))
    return {
      from: toDateString(start),
      to: toDateString(end)
    }
  })

  return {
    year,
    month,
    setYear,
    setMonth,
    range
  }
}
