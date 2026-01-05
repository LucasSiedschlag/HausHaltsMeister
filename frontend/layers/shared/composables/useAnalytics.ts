type AnalyticsPayload = Record<string, unknown>

export const useAnalytics = () => {
  const track = (event: string, payload: AnalyticsPayload = {}) => {
    if (import.meta.server) return
    const detail = { event, payload, at: new Date().toISOString() }
    window.dispatchEvent(new CustomEvent('hhm:analytics', { detail }))
  }

  return { track }
}
