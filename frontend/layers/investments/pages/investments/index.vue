<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Button } from '@shared/components/ui/button'
import ButtonGroup from '@shared/components/ui/button-group/ButtonGroup.vue'
import { Badge } from '@shared/components/ui/badge'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@shared/components/ui/card'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Skeleton } from '@shared/components/ui/skeleton'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@shared/components/ui/tooltip'
import { push } from 'notivue'
import { cn } from '@shared/utils'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { useAccounts } from '#layers/accounts/composables/useAccounts'
import { useCategories } from '#layers/categories/composables/useCategories'
import { useInvestments } from '#layers/investments/composables/useInvestments'
import { useJournalPeriod } from '#layers/journal/composables/useJournalPeriod'
import { useBalancesReport } from '#layers/reports/composables/useBalancesReport'
import { useApiClient } from '@shared/composables/useApiClient'
import { useAuth } from '#layers/auth/composables/useAuth'
import { createIdempotencyKey } from '@shared/utils/idempotency'
import { amountSchema, dateSchema, getInputClass, useInlineValidation } from '@shared/validators'
import { isApiError } from '@shared/utils/api-error'
import type { Transaction, TransactionEntry, TransactionListResponse } from '#layers/journal/composables/useJournal'

const { t, te, locale } = useI18n()
const ledgerContext = useLedgerContext()
const canEdit = computed(() => ledgerContext.hasRole('editor'))
const { setHeaderAction } = useHeaderAction()
const api = useApiClient()
const { accessToken } = useAuth()

const { accounts, fetchAccounts } = useAccounts()
const { categories, fetchCategories } = useCategories()
const { summary, loading: summaryLoading, error: summaryError, fetchSummary } = useInvestments()
const { balances, loading: balancesLoading, error: balancesError, fetchBalances } = useBalancesReport()
const journalPeriod = useJournalPeriod()

const requiredCategoryNames = ['Investimentos']

const hasActiveAccount = (type: string) =>
  accounts.value.some((account) => account.type === type && account.is_active)

const missingAccountTypes = computed(() => {
  const missing: string[] = []
  if (!hasActiveAccount('wallet') && !hasActiveAccount('current')) missing.push('cash')
  if (!hasActiveAccount('investment')) missing.push('investment')
  return missing
})

const missingCategories = computed(() =>
  requiredCategoryNames.filter((name) => !categories.value.some((category) => category.name === name))
)

const setupReady = computed(() => missingAccountTypes.value.length === 0 && missingCategories.value.length === 0)

const padMonth = (value: number) => String(value).padStart(2, '0')
const monthParam = computed(() => `${journalPeriod.year.value}-${padMonth(journalPeriod.month.value)}-01`)

const periodLabel = computed(() => {
  const date = new Date(journalPeriod.year.value, journalPeriod.month.value - 1, 1)
  return new Intl.DateTimeFormat(locale.value, { month: 'long', year: 'numeric' }).format(date)
})

const currencyFormatter = computed(() => {
  const currency = ledgerContext.activeLedger.value?.currency_code || 'BRL'
  return new Intl.NumberFormat(locale.value, { style: 'currency', currency })
})

const formatAmount = (amount: number) => currencyFormatter.value.format(amount / 100)

const summaryCards = computed(() => [
  { key: 'contributions', label: t('investments.summary.contributions'), value: summary.value?.total_contributions ?? 0 },
  { key: 'redemptions', label: t('investments.summary.redemptions'), value: summary.value?.total_redemptions ?? 0 },
  { key: 'earnings', label: t('investments.summary.earnings'), value: summary.value?.total_earnings ?? 0 },
  { key: 'losses', label: t('investments.summary.losses'), value: summary.value?.total_losses ?? 0 },
  { key: 'net', label: t('investments.summary.netVariation'), value: summary.value?.net_variation ?? 0 }
])

const netValueClass = computed(() => {
  if (!summary.value) return 'text-muted-foreground'
  if (summary.value.net_variation < 0) return 'text-destructive'
  return 'text-emerald-600 dark:text-emerald-300'
})

const formatSummaryValue = (amount: number) => (summary.value ? formatAmount(amount) : '-')

const formatCurrencyInput = (value: string) => {
  const rawDigits = value.replace(/\D/g, '')
  const digits = rawDigits.slice(0, 14)
  if (!digits) return ''
  const padded = digits.padStart(3, '0')
  const integer = padded.slice(0, -2)
  const decimal = padded.slice(-2)
  const integerValue = String(Number(integer))
  const formattedInteger = integerValue.replace(/\B(?=(\d{3})+(?!\d))/g, '.')
  return `${formattedInteger},${decimal}`
}

const parseCurrencyToCents = (value: string) => {
  const digits = value.replace(/\D/g, '').slice(0, 14)
  if (!digits) return 0
  return Number(digits)
}

const defaultDate = () => new Date().toISOString().slice(0, 10)

const amount = ref('')
const occurredAt = ref(defaultDate())
const memo = ref('')

const { touched, errors, touchField, validateAll } = useInlineValidation(
  { amount, occurredAt },
  {
    amount: amountSchema(t('investments.form.amountLabel')),
    occurredAt: dateSchema(t('investments.form.dateLabel'))
  }
)

const resetForm = () => {
  amount.value = ''
  occurredAt.value = defaultDate()
  memo.value = ''
  touched.amount = false
  touched.occurredAt = false
  errors.amount = []
  errors.occurredAt = []
}

type ActionKey = 'contribution' | 'redemption' | 'earnings' | 'loss'
const activeAction = ref<ActionKey>('contribution')
const submitting = ref(false)

const actionMeta = computed(() => ({
  title: t(`investments.forms.${activeAction.value}.title`),
  description: t(`investments.forms.${activeAction.value}.description`),
  buttonLabel: t(`investments.actions.${activeAction.value}`)
}))

const actionButtonClass = (value: ActionKey) => {
  const isActive = activeAction.value === value
  return cn(
    'h-8 px-3 text-[11px] uppercase transition-colors focus-visible:z-10 rounded-none first:rounded-l-md last:rounded-r-md border border-border/60',
    isActive
      ? 'bg-primary text-primary-foreground border-primary hover:bg-primary/90 dark:hover:bg-primary/70'
      : 'bg-background text-muted-foreground hover:bg-muted/30'
  )
}

const selectedAccountId = ref<string | null>(null)

const investmentAccounts = computed(() =>
  accounts.value
    .filter((account) => account.type === 'investment')
    .sort((a, b) => a.name.localeCompare(b.name))
)

const investmentAccountIds = computed(() => new Set(investmentAccounts.value.map((account) => account.id)))

const selectedAccount = computed(() => {
  if (!selectedAccountId.value) return null
  return investmentAccounts.value.find((account) => account.id === selectedAccountId.value) ?? null
})

const accountButtonClass = (accountId: string) => {
  const isActive = selectedAccountId.value === accountId
  return cn(
    'h-auto w-full justify-between gap-3 px-3 py-2 text-left border border-border/60 transition-colors',
    isActive
      ? 'bg-primary text-primary-foreground border-primary hover:bg-primary/90 dark:hover:bg-primary/70'
      : 'bg-background text-muted-foreground hover:bg-muted/30'
  )
}

const accountNameClass = (accountId: string) =>
  selectedAccountId.value === accountId ? 'text-primary-foreground' : 'text-foreground'

const accountTypeClass = (accountId: string) =>
  selectedAccountId.value === accountId ? 'text-primary-foreground/80' : 'text-muted-foreground'

const accountBadgeClass = (accountId: string) =>
  selectedAccountId.value === accountId ? 'border-primary-foreground/50 text-primary-foreground' : ''

const primaryCashAccount = computed(
  () =>
    accounts.value.find((account) => account.type === 'wallet' && account.is_active) ||
    accounts.value.find((account) => account.type === 'current' && account.is_active)
)

const balanceMap = computed(() => {
  const map = new Map<string, number>()
  for (const item of balances.value?.items ?? []) {
    map.set(item.account_id, item.balance_cents)
  }
  return map
})

const totalBalance = computed(() => {
  const items = balances.value?.items ?? []
  const investmentIds = investmentAccountIds.value
  return items.reduce((sum, item) => {
    if (!investmentIds.has(item.account_id)) return sum
    return sum + item.balance_cents
  }, 0)
})

const activity = ref<Transaction[]>([])
const activityLoading = ref(false)
const activityError = ref('')

const categoryById = computed(() => new Map(categories.value.map((item) => [item.id, item])))
const categoryIdByName = computed(() => new Map(categories.value.map((item) => [item.name, item.id])))

const activityItems = computed(() => {
  const accountId = selectedAccountId.value
  if (!accountId) return []
  const items = activity.value.flatMap((transaction) =>
    transaction.entries
      .filter((entry) => entry.account_id === accountId)
      .map((entry) => ({ transaction, entry }))
  )
  return items.sort((a, b) => b.transaction.occurred_at.localeCompare(a.transaction.occurred_at))
})

const formatActivityDate = (value: string) => {
  const parsed = value.includes('T') ? new Date(value) : new Date(`${value}T00:00:00`)
  if (Number.isNaN(parsed.getTime())) return value
  return new Intl.DateTimeFormat(locale.value, { day: '2-digit', month: 'short' }).format(parsed)
}

const formatActivityAmount = (entry: TransactionEntry, action?: string | null) => {
  if (action) {
    const sign = action === 'redemption' || action === 'loss' ? -1 : 1
    return formatAmount(entry.amount_cents * sign)
  }
  const category = entry.category_id ? categoryById.value.get(entry.category_id) : null
  const direction = category?.direction
  const signed = direction === 'out' ? -entry.amount_cents : entry.amount_cents
  return formatAmount(signed)
}

const activityAmountClass = (entry: TransactionEntry, action?: string | null) => {
  if (action) {
    return action === 'redemption' || action === 'loss'
      ? 'text-destructive'
      : 'text-emerald-600 dark:text-emerald-300'
  }
  const category = entry.category_id ? categoryById.value.get(entry.category_id) : null
  const direction = category?.direction
  if (direction === 'out') return 'text-destructive'
  if (direction === 'in') return 'text-emerald-600 dark:text-emerald-300'
  return 'text-foreground'
}

const refreshSummary = async () => {
  if (!ledgerContext.activeLedgerId.value || !setupReady.value) {
    summary.value = null
    return
  }
  const range = journalPeriod.range.value
  await fetchSummary(range.from, range.to)
}

const refreshBalances = async () => {
  if (!ledgerContext.activeLedgerId.value) return
  await fetchBalances(monthParam.value)
}

const fetchActivity = async () => {
  const ledgerId = ledgerContext.activeLedgerId.value
  const accountId = selectedAccountId.value
  if (!ledgerId || !accountId || !accessToken.value) {
    activity.value = []
    return
  }

  const range = journalPeriod.range.value
  const params = new URLSearchParams({
    from: range.from,
    to: range.to,
    account_id: accountId,
    limit: '60'
  })
  const path = `/ledgers/${ledgerId}/transactions?${params.toString()}`
  activityLoading.value = true
  activityError.value = ''
  try {
    const payload = await api<TransactionListResponse>(path, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      }
    })
    activity.value = payload.items
  } catch (err) {
    activityError.value = (err as Error)?.message || t('investments.activity.error')
    activity.value = []
  } finally {
    activityLoading.value = false
  }
}

const resolveErrorMessage = (err: unknown) => {
  if (isApiError(err)) {
    const key = `common.errors.${err.code}`
    if (te(key)) {
      return t(key)
    }
    return err.message || t('investments.messages.error')
  }
  return t('investments.messages.error')
}

const resolveFlowAccounts = () => {
  const selected = selectedAccount.value
  if (!selected) return null

  const cashAccount = primaryCashAccount.value
  const investmentAccount = selected
  if (!cashAccount || !investmentAccount) return null

  return { cashAccount, investmentAccount }
}

const actionDisabledReason = computed(() => {
  if (!canEdit.value) return t('investments.readOnlyHint')
  if (!setupReady.value) return t('investments.setup.blockedHint')
  if (!selectedAccount.value) return t('investments.accounts.selectHint')
  if (!selectedAccount.value.is_active) return t('investments.accounts.inactiveHint')
  if (selectedAccount.value.type !== 'investment') {
    return t('investments.accounts.unsupportedHint')
  }
  return ''
})

const actionDisabled = computed(() => Boolean(actionDisabledReason.value))

const handleSubmit = async () => {
  if (actionDisabled.value) return
  if (validateAll()) return

  const amountCents = parseCurrencyToCents(amount.value)
  if (amountCents <= 0) {
    touchField('amount')
    return
  }

  const flowAccounts = resolveFlowAccounts()
  if (!flowAccounts) {
    push.error({ title: t('investments.title'), message: t('investments.messages.error') })
    return
  }

  const categoryId = categoryIdByName.value.get('Investimentos')
  if (!categoryId) {
    push.error({ title: t('investments.title'), message: t('investments.setup.blockedHint') })
    return
  }

  const memoValue = memo.value.trim() ? memo.value.trim() : null
  const entries = [] as Array<{
    account_id: string
    category_id: string
    kind: 'transfer' | 'adjust'
    amount_cents: number
    memo?: string | null
  }>

  let description = ''
  let investmentAction: ActionKey | null = null

  try {
    if (activeAction.value === 'contribution') {
      description = 'Aporte investimentos'
      investmentAction = 'contribution'
      entries.push(
        {
          account_id: flowAccounts.cashAccount.id,
          category_id: categoryId,
          kind: 'transfer',
          amount_cents: amountCents,
          memo: memoValue
        },
        {
          account_id: flowAccounts.investmentAccount.id,
          category_id: categoryId,
          kind: 'transfer',
          amount_cents: amountCents,
          memo: memoValue
        }
      )
    }

    if (activeAction.value === 'redemption') {
      description = 'Resgate investimentos'
      investmentAction = 'redemption'
      entries.push(
        {
          account_id: flowAccounts.investmentAccount.id,
          category_id: categoryId,
          kind: 'transfer',
          amount_cents: amountCents,
          memo: memoValue
        },
        {
          account_id: flowAccounts.cashAccount.id,
          category_id: categoryId,
          kind: 'transfer',
          amount_cents: amountCents,
          memo: memoValue
        }
      )
    }

    if (activeAction.value === 'earnings') {
      if (selectedAccount.value?.type !== 'investment') {
        return
      }
      description = 'Rendimento investimentos'
      investmentAction = 'earnings'
      entries.push({
        account_id: flowAccounts.investmentAccount.id,
        category_id: categoryId,
        kind: 'adjust',
        amount_cents: amountCents,
        memo: memoValue
      })
    }

    if (activeAction.value === 'loss') {
      if (selectedAccount.value?.type !== 'investment') {
        return
      }
      description = 'Perda investimentos'
      investmentAction = 'loss'
      entries.push({
        account_id: flowAccounts.investmentAccount.id,
        category_id: categoryId,
        kind: 'adjust',
        amount_cents: amountCents,
        memo: memoValue
      })
    }
  } catch {
    push.error({ title: t('investments.title'), message: t('investments.setup.blockedHint') })
    return
  }

  const ledgerId = ledgerContext.activeLedgerId.value
  if (!ledgerId || !accessToken.value) return

  submitting.value = true
  try {
    await api<Transaction>(`/ledgers/${ledgerId}/transactions`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken.value}`
      },
      idempotencyKey: createIdempotencyKey(),
      body: {
        occurred_at: occurredAt.value,
        description,
        notes: memoValue,
        investment_action: investmentAction,
        entries
      }
    })
    push.success({ title: t('investments.title'), message: t(`investments.messages.${activeAction.value}Success`) })
    resetForm()
    await Promise.all([refreshSummary(), refreshBalances(), fetchActivity()])
  } catch (err) {
    push.error({ title: t('investments.title'), message: resolveErrorMessage(err) })
  } finally {
    submitting.value = false
  }
}

const setDefaultAccount = () => {
  const current = selectedAccountId.value
  if (current && investmentAccounts.value.some((account) => account.id === current)) {
    return
  }
  const active = investmentAccounts.value.find((account) => account.is_active)
  selectedAccountId.value = active?.id ?? investmentAccounts.value[0]?.id ?? null
}

const focusContribution = () => {
  activeAction.value = 'contribution'
  const target = document.getElementById('investment_amount')
  if (target) {
    target.scrollIntoView({ behavior: 'smooth', block: 'center' })
    ;(target as HTMLInputElement).focus()
  }
}

setHeaderAction(
  { key: 'investments:new-contribution', labelKey: 'investments.header.newContribution', requiresEditor: true },
  focusContribution
)

watch(
  () => ledgerContext.activeLedgerId.value,
  async (ledgerId) => {
    if (!ledgerId) return
    await Promise.all([fetchAccounts(), fetchCategories()])
    setDefaultAccount()
    await Promise.all([refreshSummary(), refreshBalances(), fetchActivity()])
  },
  { immediate: true }
)

watch(
  () => investmentAccounts.value.length,
  () => {
    setDefaultAccount()
  }
)

watch(
  () => journalPeriod.range.value,
  async () => {
    await Promise.all([refreshSummary(), refreshBalances(), fetchActivity()])
  }
)

watch(
  () => selectedAccountId.value,
  async () => {
    await fetchActivity()
  }
)
</script>

<template>
  <div class="grid gap-4">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
      </div>
    </div>

    <Card v-if="!setupReady" class="border-amber-500/40 bg-amber-500/10">
      <CardHeader>
        <CardTitle>{{ t('investments.setup.title') }}</CardTitle>
        <CardDescription>{{ t('investments.setup.description') }}</CardDescription>
      </CardHeader>
      <CardContent class="grid gap-4">
        <div v-if="missingAccountTypes.length" class="grid gap-2">
          <p class="text-sm font-medium">{{ t('investments.setup.accountsTitle') }}</p>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="type in missingAccountTypes" :key="type" variant="outline">
              {{ t(`investments.setup.accounts.${type}`) }}
            </Badge>
          </div>
        </div>
        <div v-if="missingCategories.length" class="grid gap-2">
          <p class="text-sm font-medium">{{ t('investments.setup.categoriesTitle') }}</p>
          <p class="text-xs text-muted-foreground">{{ t('investments.setup.categoriesHint') }}</p>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="name in missingCategories" :key="name" variant="outline">
              {{ name }}
            </Badge>
          </div>
        </div>
      </CardContent>
      <CardFooter class="flex flex-wrap gap-2">
        <Button variant="outline" as-child>
          <NuxtLink to="/accounts">{{ t('investments.setup.goAccounts') }}</NuxtLink>
        </Button>
        <Button variant="outline" as-child>
          <NuxtLink to="/categories">{{ t('investments.setup.goCategories') }}</NuxtLink>
        </Button>
      </CardFooter>
    </Card>

    <section class="grid gap-4">
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <Card v-for="card in summaryCards" :key="card.key">
          <CardHeader class="pb-2">
            <CardTitle class="text-sm font-medium text-muted-foreground">{{ card.label }}</CardTitle>
          </CardHeader>
          <CardContent>
            <Skeleton v-if="summaryLoading" class="h-7 w-24" />
            <div v-else class="text-2xl font-semibold" :class="card.key === 'net' ? netValueClass : undefined">
              {{ formatSummaryValue(card.value) }}
            </div>
          </CardContent>
        </Card>
      </div>
      <p v-if="summaryError && setupReady" class="text-sm text-destructive">
        {{ t('investments.summary.error') }}
      </p>
      <p v-if="!setupReady" class="text-xs text-muted-foreground">
        {{ t('investments.summary.unavailable') }}
      </p>
    </section>

    <section class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_450px]">
      <div class="grid gap-4">
        <Card>
        <CardHeader class="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle>{{ t('investments.balances.title') }}</CardTitle>
            <CardDescription>{{ t('investments.balances.description', { period: periodLabel }) }}</CardDescription>
          </div>
          <div class="text-right">
            <p class="text-xs text-muted-foreground">{{ t('investments.balances.totalLabel') }}</p>
            <Skeleton v-if="balancesLoading" class="mt-1 h-6 w-24" />
            <p v-else class="text-lg font-semibold">{{ formatAmount(totalBalance) }}</p>
          </div>
        </CardHeader>
        <CardContent class="grid gap-4">
          <p v-if="balancesError" class="text-xs text-destructive">{{ t('investments.balances.error') }}</p>
          <div class="grid gap-2">
            <p class="text-xs text-muted-foreground">{{ t('investments.accounts.title') }}</p>
            <div class="grid gap-2">
              <Button
                v-for="account in investmentAccounts"
                :key="account.id"
                type="button"
                variant="ghost"
                :class="accountButtonClass(account.id)"
                @click="selectedAccountId = account.id"
              >
                <div class="flex flex-col items-start">
                  <span class="text-sm font-medium" :class="accountNameClass(account.id)">{{ account.name }}</span>
                  <span class="text-xs" :class="accountTypeClass(account.id)">
                    {{ t(`accounts.types.${account.type}`) }}
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <Badge v-if="!account.is_active" variant="outline" class="text-[10px]" :class="accountBadgeClass(account.id)">
                    {{ t('investments.accounts.inactive') }}
                  </Badge>
                  <span class="text-sm font-semibold">
                    {{ balanceMap.get(account.id) !== undefined ? formatAmount(balanceMap.get(account.id) ?? 0) : '-' }}
                  </span>
                </div>
              </Button>
            </div>
          </div>
        </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{{ t('investments.activity.title') }}</CardTitle>
            <CardDescription>
              <span v-if="selectedAccount">
                {{ t('investments.activity.description', { account: selectedAccount.name, period: periodLabel }) }}
              </span>
              <span v-else>
                {{ t('investments.activity.selectAccount') }}
              </span>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div v-if="activityLoading" class="space-y-3">
              <Skeleton class="h-12 w-full" />
              <Skeleton class="h-12 w-full" />
              <Skeleton class="h-12 w-full" />
            </div>
            <p v-else-if="activityError" class="text-sm text-destructive">
              {{ t('investments.activity.error') }}
            </p>
            <div v-else class="max-h-[420px] space-y-3 overflow-y-auto pr-1">
              <div v-if="activityItems.length === 0" class="text-sm text-muted-foreground">
                {{ t('investments.activity.empty') }}
              </div>
              <div
                v-for="item in activityItems"
                :key="`${item.transaction.id}-${item.entry.id}`"
                class="flex items-start justify-between gap-3 border-b border-border/60 pb-3 last:border-b-0"
              >
                <div>
                  <p class="text-sm font-medium">{{ item.transaction.description }}</p>
                  <p class="text-xs text-muted-foreground">
                    {{ item.entry.category_id ? categoryById.get(item.entry.category_id)?.name : t('journal.entry.noCategory') }}
                  </p>
                  <p class="text-xs text-muted-foreground">
                    {{ formatActivityDate(item.transaction.occurred_at) }}
                    <span v-if="item.entry.memo">· {{ item.entry.memo }}</span>
                  </p>
                </div>
                <p class="text-sm font-semibold" :class="activityAmountClass(item.entry, item.transaction.investment_action)">
                  {{ formatActivityAmount(item.entry, item.transaction.investment_action) }}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{{ actionMeta.title }}</CardTitle>
          <CardDescription>{{ actionMeta.description }}</CardDescription>
        </CardHeader>
        <CardContent class="grid gap-4">
          <ButtonGroup class="h-8 w-fit justify-self-start">
            <Button
              type="button"
              variant="ghost"
              :class="actionButtonClass('contribution')"
              @click="activeAction = 'contribution'"
            >
              {{ t('investments.tabs.contribution') }}
            </Button>
            <Button
              type="button"
              variant="ghost"
              :class="actionButtonClass('redemption')"
              @click="activeAction = 'redemption'"
            >
              {{ t('investments.tabs.redemption') }}
            </Button>
            <Button
              type="button"
              variant="ghost"
              :class="actionButtonClass('earnings')"
              @click="activeAction = 'earnings'"
            >
              {{ t('investments.tabs.earnings') }}
            </Button>
            <Button
              type="button"
              variant="ghost"
              :class="actionButtonClass('loss')"
              @click="activeAction = 'loss'"
            >
              {{ t('investments.tabs.loss') }}
            </Button>
          </ButtonGroup>
          <div class="grid gap-3">
            <div class="grid gap-2">
              <Label>
                {{ t('investments.form.amountLabel') }}
                <span class="text-destructive">*</span>
              </Label>
              <Input
                id="investment_amount"
                :model-value="amount"
                inputmode="decimal"
                :placeholder="t('investments.form.amountPlaceholder')"
                :class="getInputClass({ touched: touched.amount, hasError: Boolean(errors.amount[0]) })"
                :disabled="actionDisabled"
                @update:model-value="(value) => { amount = formatCurrencyInput(String(value)) }"
                @blur="touchField('amount')"
              />
              <p v-if="touched.amount && errors.amount[0]" class="text-xs text-destructive">
                {{ errors.amount[0].message }}
              </p>
            </div>
            <div class="grid gap-2">
              <Label>
                {{ t('investments.form.dateLabel') }}
                <span class="text-destructive">*</span>
              </Label>
              <Input
                :model-value="occurredAt"
                type="date"
                :class="getInputClass({ touched: touched.occurredAt, hasError: Boolean(errors.occurredAt[0]) })"
                :disabled="actionDisabled"
                @update:model-value="(value) => { occurredAt = String(value) }"
                @blur="touchField('occurredAt')"
              />
              <p v-if="touched.occurredAt && errors.occurredAt[0]" class="text-xs text-destructive">
                {{ errors.occurredAt[0].message }}
              </p>
            </div>
            <div class="grid gap-2">
              <Label>{{ t('investments.form.memoLabel') }}</Label>
              <Input
                :model-value="memo"
                :placeholder="t('investments.form.memoPlaceholder')"
                :disabled="actionDisabled"
                @update:model-value="(value) => { memo = String(value) }"
              />
            </div>
          </div>
        </CardContent>
        <CardFooter class="flex justify-end">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <span>
                  <Button type="button" :disabled="actionDisabled || submitting" @click="handleSubmit">
                    {{ submitting ? t('investments.form.saving') : actionMeta.buttonLabel }}
                  </Button>
                </span>
              </TooltipTrigger>
              <TooltipContent v-if="actionDisabledReason">
                {{ actionDisabledReason }}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </CardFooter>
      </Card>
    </section>
  </div>
</template>
