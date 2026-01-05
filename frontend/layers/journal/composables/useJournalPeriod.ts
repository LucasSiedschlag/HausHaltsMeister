import { useMonthPeriod } from '@shared/composables/useMonthPeriod'

export const useJournalPeriod = () => {
  return useMonthPeriod('global')
}
