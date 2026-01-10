type HeaderAction = {
  key: string
  labelKey: string
  requiresEditor?: boolean
  disabled?: boolean
}

const handlers = new Map<string, () => void>()

export const useHeaderAction = () => {
  const action = useState<HeaderAction | null>('header_action', () => null)

  const setHeaderAction = (value: HeaderAction, handler?: () => void) => {
    const previousKey = action.value?.key
    action.value = value

    if (!import.meta.client) {
      return
    }

    if (previousKey && previousKey !== value.key) {
      handlers.delete(previousKey)
    }

    if (handler) {
      handlers.set(value.key, handler)
      return
    }

    handlers.delete(value.key)
  }

  const clearHeaderAction = (key: string) => {
    if (import.meta.client) {
      handlers.delete(key)
    }
    if (action.value?.key === key) {
      action.value = null
    }
  }

  const runHeaderAction = () => {
    const key = action.value?.key
    if (!key) return
    const handler = handlers.get(key)
    if (handler) {
      handler()
    }
  }

  const hasHandler = (key?: string | null) => Boolean(key && handlers.has(key))

  return {
    action,
    setHeaderAction,
    clearHeaderAction,
    runHeaderAction,
    hasHandler
  }
}
