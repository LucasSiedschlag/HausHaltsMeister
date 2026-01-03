<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@shared/components/ui/button'
import { Separator } from '@shared/components/ui/separator'
import { SidebarTrigger } from '@shared/components/ui/sidebar'
import LedgerSwitcher from './LedgerSwitcher.vue'
import SearchForm from './SearchForm.vue'
import { useLedgerContext } from '@shared/composables/useLedgerContext'

const { t } = useI18n()
const ledger = useLedgerContext()
const activeLedger = computed(() => ledger.activeLedger.value)
const roleLabel = computed(() => {
  const role = ledger.activeRole.value
  return role ? t(`ledgers.roles.${role}`) : t('ledgers.roles.unknown')
})
const canPostTransactions = computed(() => ledger.hasRole('editor'))
const transactionTooltip = computed(() =>
  canPostTransactions.value ? undefined : t('ledgers.roleRestrictions.transactions')
)
const exportLabel = computed(() => t('siteHeader.actions.export'))
const transactionLabel = computed(() => t('siteHeader.actions.newTransaction'))
</script>

<template>
  <header class="sticky top-0 z-30 flex h-16 items-center gap-4 border-b border-border/60 bg-background/80 px-6 backdrop-blur">
    <SidebarTrigger class="lg:hidden" />
    <Separator orientation="vertical" class="h-6 lg:hidden" />
    <div class="flex items-center gap-2">
      <LedgerSwitcher />
      <div class="flex flex-col text-xs text-muted-foreground">
        <span>{{ roleLabel }}</span>
        <span>{{ activeLedger?.currency_code }}</span>
      </div>
    </div>
    <div class="flex-1">
      <SearchForm />
    </div>
    <div class="flex items-center gap-2">
      <Button variant="outline">{{ exportLabel }}</Button>
      <Button :disabled="!canPostTransactions" :title="transactionTooltip">{{ transactionLabel }}</Button>
    </div>
  </header>
  <div v-if="ledger.error" class="border border-destructive/70 bg-destructive/10 text-destructive px-6 py-2 text-sm">
    {{ ledger.error }}
  </div>
</template>
