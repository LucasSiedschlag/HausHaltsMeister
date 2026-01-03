import { useLedgerContext } from '#layers/shared/composables/useLedgerContext'

export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/ledgers')) {
    return
  }

  const ledgerId = (to.params.ledgerId as string | undefined) ?? null
  const ledger = useLedgerContext()
  await ledger.ensureLedger(ledgerId ?? undefined)

  if (ledger.error.value) {
    return navigateTo('/auth/login')
  }
})
