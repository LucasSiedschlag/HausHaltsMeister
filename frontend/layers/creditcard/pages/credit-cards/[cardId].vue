<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Button } from '@shared/components/ui/button'
import { Card, CardContent } from '@shared/components/ui/card'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@shared/components/ui/dropdown-menu'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@shared/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@shared/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@shared/components/ui/tooltip'
import { MoreHorizontal, Plus } from 'lucide-vue-next'
import { push } from 'notivue'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { useAccounts } from '#layers/accounts/composables/useAccounts'
import { useCreditCards } from '#layers/creditcard/composables/useCreditCards'
import { useCardNetworks } from '#layers/creditcard/composables/useCardNetworks'
import { useInstallmentPlans, type InstallmentPlan } from '#layers/creditcard/composables/useInstallmentPlans'
import { useInstallments, type Installment } from '#layers/creditcard/composables/useInstallments'
import { useStatements, type Statement } from '#layers/creditcard/composables/useStatements'
import { useCategories } from '#layers/categories/composables/useCategories'
import { useJournalPeriod } from '#layers/journal/composables/useJournalPeriod'
import {
  amountSchema,
  dateSchema,
  descriptionSchema,
  integerSchema,
  monthSchema,
  nameSchema,
  getInputClass,
  useInlineValidation
} from '@shared/validators'
import { isApiError } from '@shared/utils/api-error'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'
import CreditCardTile from '#layers/creditcard/components/CreditCardTile.vue'

const { t, te, locale } = useI18n()
const route = useRoute()
const ledgerContext = useLedgerContext()
const canEdit = computed(() => ledgerContext.hasRole('editor'))
const { setHeaderAction } = useHeaderAction()

const { accounts, fetchAccounts } = useAccounts()
const { cards, fetchCard } = useCreditCards()
const { networks, fetchNetworks } = useCardNetworks()
const { categories, fetchCategories } = useCategories()
const { plans, loading: plansLoading, error: plansError, fetchPlans, createPlan, cancelPlan } = useInstallmentPlans()
const {
  installments,
  loading: installmentsLoading,
  error: installmentsError,
  fetchInstallments,
  updateInstallment,
  postMonth
} = useInstallments()
const {
  statements,
  loading: statementsLoading,
  error: statementsError,
  fetchStatements,
  closeStatement,
  payStatement
} = useStatements()
const journalPeriod = useJournalPeriod()

const cardId = computed(() => String(route.params.cardId || ''))
const cardLoading = ref(false)
const cardError = ref('')

const card = computed(() => cards.value.find((item) => item.id === cardId.value) || null)

const accountNameMap = computed(() =>
  new Map(accounts.value.map((account) => [account.id, account.name]))
)
const networkNameMap = computed(() =>
  new Map(networks.value.map((network) => [network.code, network.display_name]))
)
const categoryMap = computed(() => new Map(categories.value.map((category) => [category.id, category.name])))
const planMap = computed(() => new Map(plans.value.map((plan) => [plan.id, plan])))

const cashAccounts = computed(() => accounts.value.filter((account) => account.is_active))

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

const currencyFormatter = computed(() => {
  const currency = ledgerContext.activeLedger.value?.currency_code || 'BRL'
  return new Intl.NumberFormat(locale.value, { style: 'currency', currency })
})

const formatAmount = (amount?: number | null) => {
  if (amount === null || amount === undefined) return '-'
  return currencyFormatter.value.format(amount / 100)
}

const formatDate = (value?: string | null) => {
  if (!value) return '-'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return new Intl.DateTimeFormat(locale.value, { day: '2-digit', month: 'short', year: 'numeric' }).format(parsed)
}

const padMonth = (value: number) => String(value).padStart(2, '0')
const monthParam = computed(() => `${journalPeriod.year.value}-${padMonth(journalPeriod.month.value)}-01`)
const monthLabel = computed(() => {
  const date = new Date(journalPeriod.year.value, journalPeriod.month.value - 1, 1)
  return new Intl.DateTimeFormat(locale.value, { month: 'long', year: 'numeric' }).format(date)
})

const statusBadgeClass = (status: string) => {
  if (status === 'paid' || status === 'posted') return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
  if (status === 'closed') return 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300'
  if (status === 'cancelled' || status === 'skipped') return 'border-muted-foreground/30 bg-muted/50 text-muted-foreground'
  return 'border-border/60 text-muted-foreground'
}

const planStatusFilter = ref<'all' | 'active' | 'cancelled'>('active')
const installmentStatusFilter = ref<'all' | 'scheduled' | 'posted' | 'paid' | 'skipped'>('all')

const planFormOpen = ref(false)
const planSaving = ref(false)
const planError = ref('')

const planPurchaseDate = ref(new Date().toISOString().slice(0, 10))
const planDescription = ref('')
const planMerchant = ref('')
const planCategoryId = ref('')
const planTotalAmount = ref('')
const planInstallmentsCount = ref('')
const planFirstDueMonth = ref(`${new Date().getFullYear()}-${padMonth(new Date().getMonth() + 1)}`)

const planValidation = useInlineValidation(
  {
    planPurchaseDate,
    planDescription,
    planCategoryId,
    planTotalAmount,
    planInstallmentsCount,
    planFirstDueMonth
  },
  {
    planPurchaseDate: dateSchema(t('creditCards.plans.form.purchaseDateLabel')),
    planDescription: descriptionSchema(t('creditCards.plans.form.descriptionLabel')),
    planCategoryId: nameSchema(t('creditCards.plans.form.categoryLabel')),
    planTotalAmount: amountSchema(t('creditCards.plans.form.totalAmountLabel')),
    planInstallmentsCount: integerSchema(t('creditCards.plans.form.installmentsLabel'), { min: 1, max: 120 }),
    planFirstDueMonth: monthSchema(t('creditCards.plans.form.firstDueLabel'))
  }
)

const resetPlanForm = () => {
  planPurchaseDate.value = new Date().toISOString().slice(0, 10)
  planDescription.value = ''
  planMerchant.value = ''
  planCategoryId.value = ''
  planTotalAmount.value = ''
  planInstallmentsCount.value = ''
  planFirstDueMonth.value = `${new Date().getFullYear()}-${padMonth(new Date().getMonth() + 1)}`
  planError.value = ''
  planValidation.touched.planPurchaseDate = false
  planValidation.touched.planDescription = false
  planValidation.touched.planCategoryId = false
  planValidation.touched.planTotalAmount = false
  planValidation.touched.planInstallmentsCount = false
  planValidation.touched.planFirstDueMonth = false
  planValidation.errors.planPurchaseDate = []
  planValidation.errors.planDescription = []
  planValidation.errors.planCategoryId = []
  planValidation.errors.planTotalAmount = []
  planValidation.errors.planInstallmentsCount = []
  planValidation.errors.planFirstDueMonth = []
}

const openPlanForm = () => {
  resetPlanForm()
  planFormOpen.value = true
}

const handleCreatePlan = async () => {
  planError.value = ''
  if (planValidation.validateAll()) return
  const totalCents = parseCurrencyToCents(planTotalAmount.value)
  const count = Number(planInstallmentsCount.value)
  const installmentAmount = Math.floor(totalCents / count)
  if (!count || installmentAmount <= 0) {
    planError.value = t('creditCards.plans.messages.installmentTooSmall')
    return
  }
  planSaving.value = true
  try {
    await createPlan(cardId.value, {
      purchase_occurred_at: planPurchaseDate.value,
      merchant: planMerchant.value.trim() || null,
      description: planDescription.value.trim(),
      category_id: planCategoryId.value,
      total_amount_cents: totalCents,
      installments_count: count,
      installment_amount_cents: installmentAmount,
      first_due_month: `${planFirstDueMonth.value}-01`
    })
    push.success({ title: t('creditCards.plans.title'), message: t('creditCards.plans.messages.created') })
    planFormOpen.value = false
    await refreshPlans()
  } catch (err) {
    if (isApiError(err) && te(`common.errors.${err.code}`)) {
      planError.value = t(`common.errors.${err.code}`)
    } else {
      planError.value = t('creditCards.plans.messages.error')
    }
  } finally {
    planSaving.value = false
  }
}

const confirmPlanOpen = ref(false)
const confirmPlan = ref<InstallmentPlan | null>(null)

const requestCancelPlan = (plan: InstallmentPlan) => {
  confirmPlan.value = plan
  confirmPlanOpen.value = true
}

const handleCancelPlan = async () => {
  if (!confirmPlan.value) return
  try {
    await cancelPlan(cardId.value, confirmPlan.value.id)
    push.success({ title: t('creditCards.plans.title'), message: t('creditCards.plans.messages.cancelled') })
  } catch (err) {
    if (isApiError(err) && te(`common.errors.${err.code}`)) {
      push.error({ title: t('creditCards.plans.title'), message: t(`common.errors.${err.code}`) })
    } else {
      push.error({ title: t('creditCards.plans.title'), message: t('creditCards.plans.messages.error') })
    }
  } finally {
    confirmPlan.value = null
  }
}

const confirmInstallmentOpen = ref(false)
const confirmInstallment = ref<Installment | null>(null)

const requestSkipInstallment = (installment: Installment) => {
  confirmInstallment.value = installment
  confirmInstallmentOpen.value = true
}

const handleSkipInstallment = async () => {
  if (!confirmInstallment.value) return
  try {
    await updateInstallment(cardId.value, confirmInstallment.value.id, 'skipped')
    push.success({ title: t('creditCards.installments.title'), message: t('creditCards.installments.messages.skipped') })
  } catch (err) {
    if (isApiError(err) && te(`common.errors.${err.code}`)) {
      push.error({ title: t('creditCards.installments.title'), message: t(`common.errors.${err.code}`) })
    } else {
      push.error({ title: t('creditCards.installments.title'), message: t('creditCards.installments.messages.error') })
    }
  } finally {
    confirmInstallment.value = null
  }
}

const postLoading = ref(false)

const handlePostMonth = async () => {
  if (!cardId.value) return
  postLoading.value = true
  try {
    const result = await postMonth(cardId.value, monthParam.value)
    if (result?.posted_count === 0) {
      push.success({ title: t('creditCards.installments.title'), message: t('creditCards.messages.postEmpty') })
    } else {
      push.success({ title: t('creditCards.installments.title'), message: t('creditCards.messages.postSuccess', { count: result?.posted_count || 0 }) })
    }
    await refreshInstallments()
  } catch (err) {
    if (isApiError(err) && te(`common.errors.${err.code}`)) {
      push.error({ title: t('creditCards.installments.title'), message: t(`common.errors.${err.code}`) })
    } else {
      push.error({ title: t('creditCards.installments.title'), message: t('creditCards.messages.error') })
    }
  } finally {
    postLoading.value = false
  }
}

const closeLoading = ref(false)

const handleCloseStatement = async () => {
  closeLoading.value = true
  try {
    await closeStatement(cardId.value, monthParam.value)
    push.success({ title: t('creditCards.statements.title'), message: t('creditCards.statements.messages.closed') })
    await refreshStatements()
  } catch (err) {
    if (isApiError(err) && te(`common.errors.${err.code}`)) {
      push.error({ title: t('creditCards.statements.title'), message: t(`common.errors.${err.code}`) })
    } else {
      push.error({ title: t('creditCards.statements.title'), message: t('creditCards.statements.messages.error') })
    }
  } finally {
    closeLoading.value = false
  }
}

const payFormOpen = ref(false)
const paySaving = ref(false)
const payError = ref('')

const payStatementId = ref('')
const payAmount = ref('')
const payDate = ref(new Date().toISOString().slice(0, 10))
const payCashAccountId = ref('')

const payValidation = useInlineValidation(
  { payStatementId, payAmount, payDate, payCashAccountId },
  {
    payStatementId: nameSchema(t('creditCards.statements.form.statementLabel')),
    payAmount: amountSchema(t('creditCards.statements.form.amountLabel')),
    payDate: dateSchema(t('creditCards.statements.form.dateLabel')),
    payCashAccountId: nameSchema(t('creditCards.statements.form.cashAccountLabel'))
  }
)

const resetPayForm = () => {
  payStatementId.value = ''
  payAmount.value = ''
  payDate.value = new Date().toISOString().slice(0, 10)
  payCashAccountId.value = cashAccounts.value[0]?.id ?? ''
  payError.value = ''
  payValidation.touched.payStatementId = false
  payValidation.touched.payAmount = false
  payValidation.touched.payDate = false
  payValidation.touched.payCashAccountId = false
  payValidation.errors.payStatementId = []
  payValidation.errors.payAmount = []
  payValidation.errors.payDate = []
  payValidation.errors.payCashAccountId = []
}

const currentStatement = computed(() => {
  if (!statements.value.length) return null
  return statements.value[0]
})

const openPayForm = (statement?: Statement) => {
  resetPayForm()
  const target = statement || currentStatement.value
  if (target) {
    payStatementId.value = target.id
    const remaining = Math.max(0, target.total_charges_cents - target.total_payments_cents)
    payAmount.value = formatCurrencyInput(String(remaining))
  }
  payFormOpen.value = true
}

const handlePayStatement = async () => {
  payError.value = ''
  if (payValidation.validateAll()) return
  const amountCents = parseCurrencyToCents(payAmount.value)
  paySaving.value = true
  try {
    await payStatement(cardId.value, {
      statement_id: payStatementId.value,
      payment_date: payDate.value,
      pay_amount_cents: amountCents,
      paying_account_id: payCashAccountId.value
    })
    push.success({ title: t('creditCards.statements.title'), message: t('creditCards.statements.messages.paid') })
    payFormOpen.value = false
    await refreshStatements()
  } catch (err) {
    if (isApiError(err) && te(`common.errors.${err.code}`)) {
      payError.value = t(`common.errors.${err.code}`)
    } else {
      payError.value = t('creditCards.statements.messages.error')
    }
  } finally {
    paySaving.value = false
  }
}

const eligibleCategories = computed(() =>
  categories.value
    .filter((category) => category.direction === 'out' && category.is_budget_relevant && category.is_active)
    .sort((a, b) => a.name.localeCompare(b.name))
)

const statementSummary = computed(() => {
  if (!currentStatement.value) return null
  const remaining = Math.max(0, currentStatement.value.total_charges_cents - currentStatement.value.total_payments_cents)
  return {
    remaining,
    status: currentStatement.value.status
  }
})

const refreshPlans = async () => {
  if (!cardId.value) return
  const status = planStatusFilter.value === 'all' ? undefined : planStatusFilter.value
  await fetchPlans(cardId.value, status)
}

const refreshInstallments = async () => {
  if (!cardId.value) return
  const status = installmentStatusFilter.value === 'all' ? undefined : installmentStatusFilter.value
  await fetchInstallments(cardId.value, { month: monthParam.value, status })
}

const refreshStatements = async () => {
  if (!cardId.value) return
  await fetchStatements(cardId.value, monthParam.value)
}

const loadCard = async () => {
  if (!cardId.value) return
  cardLoading.value = true
  cardError.value = ''
  try {
    await fetchCard(cardId.value)
  } catch (err) {
    cardError.value = (err as Error)?.message || t('creditCards.messages.error')
  } finally {
    cardLoading.value = false
  }
}

setHeaderAction(
  { key: 'credit-cards:new-plan', labelKey: 'creditCards.header.newPlan', requiresEditor: true },
  openPlanForm
)

watch(
  () => ledgerContext.activeLedgerId.value,
  async (ledgerId) => {
    if (!ledgerId) return
    await Promise.all([fetchAccounts(), fetchNetworks(), fetchCategories({ direction: 'out', active: true })])
    await loadCard()
    await Promise.all([refreshPlans(), refreshInstallments(), refreshStatements()])
  },
  { immediate: true }
)

watch(
  () => cardId.value,
  async () => {
    if (!cardId.value) return
    await loadCard()
    await Promise.all([refreshPlans(), refreshInstallments(), refreshStatements()])
  }
)

watch(
  () => planStatusFilter.value,
  async () => {
    await refreshPlans()
  }
)

watch(
  () => [installmentStatusFilter.value, monthParam.value],
  async () => {
    await refreshInstallments()
  }
)

watch(
  () => monthParam.value,
  async () => {
    await refreshStatements()
  }
)
</script>

<template>
  <div class="grid gap-6">
    <div class="grid gap-6 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
      <CreditCardTile
        v-if="card"
        class="w-full"
        :card="card"
        :account-name="accountNameMap.get(card.parent_account_id) || card.parent_account_id"
        :network-name="networkNameMap.get(card.brand) || card.brand"
        :clickable="false"
        :show-actions="false"
      />
      <div v-else class="text-sm text-muted-foreground">
        {{ t('creditCards.detail.unknownCard') }}
      </div>

      <Card>
        <CardContent class="space-y-4 pt-6">
          <div class="flex items-start gap-4 text-sm">
            <div>
              <p class="text-xs uppercase text-muted-foreground">{{ t('creditCards.detail.statementMonth') }}</p>
              <p class="font-medium">{{ monthLabel }}</p>
            </div>
            <span class="mt-4 text-xs text-muted-foreground">|</span>
            <div>
              <p class="text-xs uppercase text-muted-foreground">{{ t('creditCards.detail.statementStatus') }}</p>
              <div class="flex items-center gap-2">
                <span
                  class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs"
                  :class="statusBadgeClass(statementSummary?.status || 'open')"
                >
                  {{ t(`creditCards.statements.status.${statementSummary?.status || 'open'}`) }}
                </span>
                <span class="text-xs text-muted-foreground">
                  {{ formatAmount(statementSummary?.remaining ?? null) }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <span>
                    <Button :disabled="!canEdit || postLoading" @click="handlePostMonth">
                      {{ t('creditCards.actions.postMonth') }}
                    </Button>
                  </span>
                </TooltipTrigger>
                <TooltipContent v-if="!canEdit">
                  {{ t('creditCards.readOnlyHint') }}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <span>
                    <Button variant="outline" :disabled="!canEdit || closeLoading" @click="handleCloseStatement">
                      {{ t('creditCards.actions.closeStatement') }}
                    </Button>
                  </span>
                </TooltipTrigger>
                <TooltipContent v-if="!canEdit">
                  {{ t('creditCards.readOnlyHint') }}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <span>
                    <Button variant="secondary" :disabled="!canEdit" @click="openPayForm()">
                      {{ t('creditCards.actions.payStatement') }}
                    </Button>
                  </span>
                </TooltipTrigger>
                <TooltipContent v-if="!canEdit">
                  {{ t('creditCards.readOnlyHint') }}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </CardContent>
      </Card>
    </div>

    <Tabs default-value="plans" class="space-y-4">
      <TabsList class="flex w-full flex-wrap justify-start gap-2">
        <TabsTrigger value="plans">{{ t('creditCards.plans.title') }}</TabsTrigger>
        <TabsTrigger value="installments">{{ t('creditCards.installments.title') }}</TabsTrigger>
        <TabsTrigger value="statements">{{ t('creditCards.statements.title') }}</TabsTrigger>
      </TabsList>

      <TabsContent value="plans" class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="w-full max-w-xs">
            <Label class="text-xs text-muted-foreground">{{ t('creditCards.plans.filters.status') }}</Label>
            <Select v-model="planStatusFilter">
              <SelectTrigger class="mt-2">
                <SelectValue :placeholder="t('creditCards.plans.filters.placeholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{{ t('creditCards.plans.filters.all') }}</SelectItem>
                <SelectItem value="active">{{ t('creditCards.plans.filters.active') }}</SelectItem>
                <SelectItem value="cancelled">{{ t('creditCards.plans.filters.cancelled') }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <span>
                  <Button :disabled="!canEdit" @click="openPlanForm">
                    <Plus class="h-4 w-4" />
                    {{ t('creditCards.actions.newPlan') }}
                  </Button>
                </span>
              </TooltipTrigger>
              <TooltipContent v-if="!canEdit">
                {{ t('creditCards.readOnlyHint') }}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        <Card>
          <CardContent class="pt-6">
            <div v-if="plansLoading" class="text-sm text-muted-foreground">
              {{ t('creditCards.plans.loading') }}
            </div>
            <div v-else-if="plansError" class="text-sm text-destructive">
              {{ plansError }}
            </div>
            <div v-else-if="plans.length === 0" class="text-sm text-muted-foreground">
              {{ t('creditCards.plans.empty') }}
            </div>
            <div v-else class="overflow-hidden rounded-lg border">
              <table class="w-full text-sm">
                <thead class="bg-muted/50 text-xs uppercase text-muted-foreground">
                  <tr>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.plans.table.description') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.plans.table.category') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.plans.table.amount') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.plans.table.firstDue') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.plans.table.status') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('creditCards.plans.table.actions') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="plan in plans" :key="plan.id" class="border-t">
                    <td class="px-4 py-3">
                      <div class="font-medium">{{ plan.description }}</div>
                      <div class="text-xs text-muted-foreground">
                        {{ plan.merchant || t('creditCards.plans.table.noMerchant') }}
                      </div>
                    </td>
                    <td class="px-4 py-3 text-xs text-muted-foreground">
                      {{ categoryMap.get(plan.category_id) || plan.category_id }}
                    </td>
                    <td class="px-4 py-3">
                      {{ formatAmount(plan.total_amount_cents) }}
                      <div class="text-xs text-muted-foreground">
                        {{ plan.installments_count }}x {{ formatAmount(plan.installment_amount_cents) }}
                      </div>
                    </td>
                    <td class="px-4 py-3 text-xs text-muted-foreground">
                      {{ formatDate(plan.first_due_month) }}
                    </td>
                    <td class="px-4 py-3">
                      <span class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs" :class="statusBadgeClass(plan.status)">
                        {{ t(`creditCards.plans.status.${plan.status}`) }}
                      </span>
                    </td>
                    <td class="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger as-child>
                          <Button variant="ghost" size="icon">
                            <MoreHorizontal class="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem :disabled="!canEdit || plan.status !== 'active'" @click="requestCancelPlan(plan)">
                            {{ t('creditCards.actions.cancelPlan') }}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="installments" class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="w-full max-w-xs">
            <Label class="text-xs text-muted-foreground">{{ t('creditCards.installments.filters.status') }}</Label>
            <Select v-model="installmentStatusFilter">
              <SelectTrigger class="mt-2">
                <SelectValue :placeholder="t('creditCards.installments.filters.placeholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{{ t('creditCards.installments.filters.all') }}</SelectItem>
                <SelectItem value="scheduled">{{ t('creditCards.installments.filters.scheduled') }}</SelectItem>
                <SelectItem value="posted">{{ t('creditCards.installments.filters.posted') }}</SelectItem>
                <SelectItem value="paid">{{ t('creditCards.installments.filters.paid') }}</SelectItem>
                <SelectItem value="skipped">{{ t('creditCards.installments.filters.skipped') }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="text-xs text-muted-foreground">
            {{ t('creditCards.installments.period', { month: monthLabel }) }}
          </div>
        </div>

        <Card>
          <CardContent class="pt-6">
            <div v-if="installmentsLoading" class="text-sm text-muted-foreground">
              {{ t('creditCards.installments.loading') }}
            </div>
            <div v-else-if="installmentsError" class="text-sm text-destructive">
              {{ installmentsError }}
            </div>
            <div v-else-if="installments.length === 0" class="text-sm text-muted-foreground">
              {{ t('creditCards.installments.empty') }}
            </div>
            <div v-else class="overflow-hidden rounded-lg border">
              <table class="w-full text-sm">
                <thead class="bg-muted/50 text-xs uppercase text-muted-foreground">
                  <tr>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.installments.table.plan') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.installments.table.amount') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.installments.table.dueMonth') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.installments.table.status') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('creditCards.installments.table.actions') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="installment in installments" :key="installment.id" class="border-t">
                    <td class="px-4 py-3">
                      <div class="font-medium">
                        {{ planMap.get(installment.plan_id)?.description || installment.plan_id }}
                      </div>
                      <div class="text-xs text-muted-foreground">
                        {{ t('creditCards.installments.table.installmentLabel', {
                          number: installment.installment_no,
                          total: planMap.get(installment.plan_id)?.installments_count || '-'
                        }) }}
                      </div>
                    </td>
                    <td class="px-4 py-3">{{ formatAmount(installment.amount_cents) }}</td>
                    <td class="px-4 py-3 text-xs text-muted-foreground">{{ formatDate(installment.due_month) }}</td>
                    <td class="px-4 py-3">
                      <span class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs" :class="statusBadgeClass(installment.status)">
                        {{ t(`creditCards.installments.status.${installment.status}`) }}
                      </span>
                    </td>
                    <td class="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger as-child>
                          <Button variant="ghost" size="icon">
                            <MoreHorizontal class="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            :disabled="!canEdit || installment.status !== 'scheduled'"
                            @click="requestSkipInstallment(installment)"
                          >
                            {{ t('creditCards.actions.skipInstallment') }}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="statements" class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="text-xs text-muted-foreground">
            {{ t('creditCards.statements.period', { month: monthLabel }) }}
          </div>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <span>
                  <Button variant="outline" :disabled="!canEdit" @click="openPayForm()">
                    {{ t('creditCards.actions.payStatement') }}
                  </Button>
                </span>
              </TooltipTrigger>
              <TooltipContent v-if="!canEdit">
                {{ t('creditCards.readOnlyHint') }}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        <Card>
          <CardContent class="pt-6">
            <div v-if="statementsLoading" class="text-sm text-muted-foreground">
              {{ t('creditCards.statements.loading') }}
            </div>
            <div v-else-if="statementsError" class="text-sm text-destructive">
              {{ statementsError }}
            </div>
            <div v-else-if="statements.length === 0" class="text-sm text-muted-foreground">
              {{ t('creditCards.statements.empty') }}
            </div>
            <div v-else class="overflow-hidden rounded-lg border">
              <table class="w-full text-sm">
                <thead class="bg-muted/50 text-xs uppercase text-muted-foreground">
                  <tr>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.statements.table.month') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.statements.table.closing') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.statements.table.due') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.statements.table.total') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.statements.table.payments') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('creditCards.statements.table.status') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('creditCards.statements.table.actions') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="statement in statements" :key="statement.id" class="border-t">
                    <td class="px-4 py-3 text-xs text-muted-foreground">
                      {{ formatDate(statement.statement_month) }}
                    </td>
                    <td class="px-4 py-3 text-xs text-muted-foreground">
                      {{ formatDate(statement.closing_date) }}
                    </td>
                    <td class="px-4 py-3 text-xs text-muted-foreground">
                      {{ formatDate(statement.due_date) }}
                    </td>
                    <td class="px-4 py-3">{{ formatAmount(statement.total_charges_cents) }}</td>
                    <td class="px-4 py-3">{{ formatAmount(statement.total_payments_cents) }}</td>
                    <td class="px-4 py-3">
                      <span class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs" :class="statusBadgeClass(statement.status)">
                        {{ t(`creditCards.statements.status.${statement.status}`) }}
                      </span>
                    </td>
                    <td class="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger as-child>
                          <Button variant="ghost" size="icon">
                            <MoreHorizontal class="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem :disabled="!canEdit" @click="openPayForm(statement)">
                            {{ t('creditCards.actions.payStatement') }}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <Sheet v-model:open="planFormOpen">
      <SheetContent class="sm:max-w-xl">
        <SheetHeader>
          <SheetTitle>{{ t('creditCards.plans.form.title') }}</SheetTitle>
          <SheetDescription>{{ t('creditCards.plans.form.description') }}</SheetDescription>
        </SheetHeader>

        <div class="mt-6 space-y-4">
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>{{ t('creditCards.plans.form.purchaseDateLabel') }}</Label>
              <Input
                v-model="planPurchaseDate"
                type="date"
                :class="getInputClass({ touched: planValidation.touched.planPurchaseDate, hasError: Boolean(planValidation.errors.planPurchaseDate?.[0]) })"
                @blur="planValidation.touchField('planPurchaseDate')"
              />
              <p v-if="planValidation.touched.planPurchaseDate && planValidation.errors.planPurchaseDate.length" class="text-xs text-destructive">
                {{ planValidation.errors.planPurchaseDate?.[0]?.message }}
              </p>
            </div>
            <div class="space-y-2">
              <Label>{{ t('creditCards.plans.form.firstDueLabel') }}</Label>
              <Input
                v-model="planFirstDueMonth"
                type="month"
                :class="getInputClass({ touched: planValidation.touched.planFirstDueMonth, hasError: Boolean(planValidation.errors.planFirstDueMonth?.[0]) })"
                @blur="planValidation.touchField('planFirstDueMonth')"
              />
              <p v-if="planValidation.touched.planFirstDueMonth && planValidation.errors.planFirstDueMonth.length" class="text-xs text-destructive">
                {{ planValidation.errors.planFirstDueMonth?.[0]?.message }}
              </p>
            </div>
          </div>

          <div class="space-y-2">
            <Label>{{ t('creditCards.plans.form.descriptionLabel') }}</Label>
            <Input
              v-model="planDescription"
              :placeholder="t('creditCards.plans.form.descriptionPlaceholder')"
              :class="getInputClass({ touched: planValidation.touched.planDescription, hasError: Boolean(planValidation.errors.planDescription?.[0]) })"
              @blur="planValidation.touchField('planDescription')"
            />
            <p v-if="planValidation.touched.planDescription && planValidation.errors.planDescription.length" class="text-xs text-destructive">
              {{ planValidation.errors.planDescription?.[0]?.message }}
            </p>
          </div>

          <div class="space-y-2">
            <Label>{{ t('creditCards.plans.form.merchantLabel') }}</Label>
            <Input v-model="planMerchant" :placeholder="t('creditCards.plans.form.merchantPlaceholder')" />
          </div>

          <div class="space-y-2">
            <Label>{{ t('creditCards.plans.form.categoryLabel') }}</Label>
            <Select v-model="planCategoryId">
            <SelectTrigger :class="getInputClass({ touched: planValidation.touched.planCategoryId, hasError: Boolean(planValidation.errors.planCategoryId?.[0]) })">
                <SelectValue :placeholder="t('creditCards.plans.form.categoryPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="category in eligibleCategories" :key="category.id" :value="category.id">
                  {{ category.name }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="planValidation.touched.planCategoryId && planValidation.errors.planCategoryId.length" class="text-xs text-destructive">
              {{ planValidation.errors.planCategoryId?.[0]?.message }}
            </p>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>{{ t('creditCards.plans.form.totalAmountLabel') }}</Label>
              <Input
                :model-value="planTotalAmount"
                inputmode="numeric"
                :placeholder="t('creditCards.plans.form.totalAmountPlaceholder')"
                :class="getInputClass({ touched: planValidation.touched.planTotalAmount, hasError: Boolean(planValidation.errors.planTotalAmount?.[0]) })"
                @update:model-value="(value) => { planTotalAmount = formatCurrencyInput(String(value)) }"
                @blur="planValidation.touchField('planTotalAmount')"
              />
              <p v-if="planValidation.touched.planTotalAmount && planValidation.errors.planTotalAmount.length" class="text-xs text-destructive">
                {{ planValidation.errors.planTotalAmount?.[0]?.message }}
              </p>
            </div>
            <div class="space-y-2">
              <Label>{{ t('creditCards.plans.form.installmentsLabel') }}</Label>
              <Input
                v-model="planInstallmentsCount"
                inputmode="numeric"
                :placeholder="t('creditCards.plans.form.installmentsPlaceholder')"
                :class="getInputClass({ touched: planValidation.touched.planInstallmentsCount, hasError: Boolean(planValidation.errors.planInstallmentsCount?.[0]) })"
                @blur="planValidation.touchField('planInstallmentsCount')"
              />
              <p v-if="planValidation.touched.planInstallmentsCount && planValidation.errors.planInstallmentsCount.length" class="text-xs text-destructive">
                {{ planValidation.errors.planInstallmentsCount?.[0]?.message }}
              </p>
            </div>
          </div>

          <p v-if="planError" class="text-sm text-destructive">
            {{ planError }}
          </p>
        </div>

        <SheetFooter class="mt-6">
          <Button variant="outline" @click="planFormOpen = false">
            {{ t('creditCards.form.cancel') }}
          </Button>
          <Button :disabled="planSaving" @click="handleCreatePlan">
            {{ planSaving ? t('creditCards.form.saving') : t('creditCards.plans.form.submit') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <Sheet v-model:open="payFormOpen">
      <SheetContent class="sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>{{ t('creditCards.statements.form.title') }}</SheetTitle>
          <SheetDescription>{{ t('creditCards.statements.form.description') }}</SheetDescription>
        </SheetHeader>

        <div class="mt-6 space-y-4">
          <div class="space-y-2">
            <Label>{{ t('creditCards.statements.form.statementLabel') }}</Label>
            <Select v-model="payStatementId">
            <SelectTrigger :class="getInputClass({ touched: payValidation.touched.payStatementId, hasError: Boolean(payValidation.errors.payStatementId?.[0]) })">
                <SelectValue :placeholder="t('creditCards.statements.form.statementPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="statement in statements" :key="statement.id" :value="statement.id">
                  {{ formatDate(statement.statement_month) }} · {{ formatAmount(statement.total_charges_cents) }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="payValidation.touched.payStatementId && payValidation.errors.payStatementId.length" class="text-xs text-destructive">
              {{ payValidation.errors.payStatementId?.[0]?.message }}
            </p>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>{{ t('creditCards.statements.form.amountLabel') }}</Label>
              <Input
                :model-value="payAmount"
                inputmode="numeric"
                :placeholder="t('creditCards.statements.form.amountPlaceholder')"
                :class="getInputClass({ touched: payValidation.touched.payAmount, hasError: Boolean(payValidation.errors.payAmount?.[0]) })"
                @update:model-value="(value) => { payAmount = formatCurrencyInput(String(value)) }"
                @blur="payValidation.touchField('payAmount')"
              />
              <p v-if="payValidation.touched.payAmount && payValidation.errors.payAmount.length" class="text-xs text-destructive">
                {{ payValidation.errors.payAmount?.[0]?.message }}
              </p>
            </div>
            <div class="space-y-2">
              <Label>{{ t('creditCards.statements.form.dateLabel') }}</Label>
              <Input
                v-model="payDate"
                type="date"
                :class="getInputClass({ touched: payValidation.touched.payDate, hasError: Boolean(payValidation.errors.payDate?.[0]) })"
                @blur="payValidation.touchField('payDate')"
              />
              <p v-if="payValidation.touched.payDate && payValidation.errors.payDate.length" class="text-xs text-destructive">
                {{ payValidation.errors.payDate?.[0]?.message }}
              </p>
            </div>
          </div>

          <div class="space-y-2">
            <Label>{{ t('creditCards.statements.form.cashAccountLabel') }}</Label>
            <Select v-model="payCashAccountId">
            <SelectTrigger :class="getInputClass({ touched: payValidation.touched.payCashAccountId, hasError: Boolean(payValidation.errors.payCashAccountId?.[0]) })">
                <SelectValue :placeholder="t('creditCards.statements.form.cashAccountPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="account in cashAccounts" :key="account.id" :value="account.id">
                  {{ account.name }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="payValidation.touched.payCashAccountId && payValidation.errors.payCashAccountId.length" class="text-xs text-destructive">
              {{ payValidation.errors.payCashAccountId?.[0]?.message }}
            </p>
          </div>

          <p v-if="payError" class="text-sm text-destructive">
            {{ payError }}
          </p>
        </div>

        <SheetFooter class="mt-6">
          <Button variant="outline" @click="payFormOpen = false">
            {{ t('creditCards.form.cancel') }}
          </Button>
          <Button :disabled="paySaving" @click="handlePayStatement">
            {{ paySaving ? t('creditCards.form.saving') : t('creditCards.statements.form.submit') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <ConfirmDialog
      v-model:open="confirmPlanOpen"
      :title="t('creditCards.plans.confirm.deleteTitle')"
      :description="t('creditCards.plans.confirm.deleteMessage', { name: confirmPlan?.description || '' })"
      :confirm-label="t('creditCards.plans.confirm.confirmAction')"
      :cancel-label="t('creditCards.plans.confirm.cancelAction')"
      @confirm="handleCancelPlan"
      @cancel="confirmPlan = null"
    />

    <ConfirmDialog
      v-model:open="confirmInstallmentOpen"
      :title="t('creditCards.installments.confirm.skipTitle')"
      :description="t('creditCards.installments.confirm.skipMessage')"
      :confirm-label="t('creditCards.installments.confirm.confirmAction')"
      :cancel-label="t('creditCards.installments.confirm.cancelAction')"
      @confirm="handleSkipInstallment"
      @cancel="confirmInstallment = null"
    />
  </div>
</template>
