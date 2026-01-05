export const useJournalUi = () => {
  const createOpen = useState<boolean>('journal_create_open', () => false)

  const openCreate = () => {
    createOpen.value = true
  }

  return {
    createOpen,
    openCreate
  }
}
