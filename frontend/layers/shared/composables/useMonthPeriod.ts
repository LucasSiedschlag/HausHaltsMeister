import { computed } from 'vue'

const toDateString = (date: Date) => date.toISOString().slice(0, 10)

export const useMonthPeriod = (contextId: string = 'global') => {

  const now = new Date()
  const year = useCookie<number>(`period_year_${contextId}`, { default: () => now.getFullYear() })
  const month = useCookie<number>(`period_month_${contextId}`, { default: () => now.getMonth() + 1 })

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

  const nextMonth = () => {
    if (month.value === 12) {
      month.value = 1
      year.value++
    } else {
      month.value++
    }
  }

  const previousMonth = () => {
    if (month.value === 1) {
      month.value = 12
      year.value--
    } else {
      month.value--
    }
  }

  return {
    year,
    month,
    setYear,
    setMonth,
    nextMonth,
    previousMonth,
    range
  }
}
