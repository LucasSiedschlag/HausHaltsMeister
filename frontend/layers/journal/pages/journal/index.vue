<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import type { ColumnDef, SortingState, VisibilityState } from '@tanstack/vue-table'
import { FlexRender, getCoreRowModel, getSortedRowModel, useVueTable } from '@tanstack/vue-table'
import { Button } from '@shared/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@shared/components/ui/dropdown-menu'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@shared/components/ui/sheet'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@shared/components/ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@shared/components/ui/tooltip'
import { CalendarIcon, ChevronDown, MoreHorizontal, Plus, Trash2 } from 'lucide-vue-next'
import { push } from 'notivue'
import CrudTableCard from '@shared/components/CrudTableCard.vue'
import SortableColumnHeader from '@shared/components/SortableColumnHeader.vue'

import { Calendar } from '@shared/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@shared/components/ui/popover'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { useAccounts } from '#layers/accounts/composables/useAccounts'
import { useCategories } from '#layers/categories/composables/useCategories'
import { useJournal, type EntryKind, type Transaction, type TransactionEntryInput } from '#layers/journal/composables/useJournal'
import { useJournalUi } from '#layers/journal/composables/useJournalUi'
import { useJournalPeriod } from '#layers/journal/composables/useJournalPeriod'
import { dateSchema, descriptionSchema, getInputClass, useInlineValidation } from '@shared/validators'
import { isApiError } from '@shared/utils/api-error'
import { getLocalTimeZone, parseDate } from '@internationalized/date'

const { t, locale } = useI18n()
const ledgerContext = useLedgerContext()
const canEdit = computed(() => ledgerContext.hasRole('editor'))

const { accounts, fetchAccounts } = useAccounts()
const { categories, fetchCategories } = useCategories()
const {
  transactions,
  nextCursor,
  lastPageSize,
  loading,
  error,
  fetchTransactions,
  createTransaction,
  updateTransaction
} = useJournal()

const journalUi = useJournalUi()
const journalPeriod = useJournalPeriod()
const { setHeaderAction } = useHeaderAction()

setHeaderAction(
  { key: 'journal:new-transaction', labelKey: 'journal.header.newTransaction', requiresEditor: true },
  journalUi.openCreate
)

const filterQuery = ref('')
const filterAccounts = ref<string[]>([])
const filterCategories = ref<string[]>([])
const filterLimit = ref('50')

const appliedFilters = ref({
  q: '',
  from: '',
  to: '',
  account_ids: [] as string[],
  category_ids: [] as string[],
  limit: 50
})

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editingTransaction = ref<Transaction | null>(null)
const isSaving = ref(false)
const formError = ref('')

const occurredAt = ref('')
const description = ref('')
const notes = ref('')
const calendarDate = ref<unknown>(undefined)

type EntryDraft = {
  id: string
  account_id: string
  category_id: string
  kind: EntryKind
  amount_cents: string
  memo: string
}

type EntryDraftError = {
  account?: string
  amount?: string
  kind?: string
}

type JournalRow = {
  row_id: string
  transaction_id: string
  occurred_at: string
  description: string
  notes?: string | null
  account_id: string
  category_id?: string | null
  account_name: string
  category_name: string
  kind: EntryKind
  amount_cents: number
  memo?: string | null
}

const entries = ref<EntryDraft[]>([])
const entryErrors = ref<EntryDraftError[]>([])

const valueUpdater = <T>(updaterOrValue: T | ((previous: T) => T), target: { value: T }) => {
  target.value = typeof updaterOrValue === 'function'
    ? (updaterOrValue as (previous: T) => T)(target.value)
    : updaterOrValue
}

const normalizeCalendarValue = (value: unknown) => {
  if (!value) return null
  if (Array.isArray(value)) {
    return value[0] ?? null
  }
  return value
}

const formatCalendarLabel = (value: unknown) => {
  const normalized = normalizeCalendarValue(value)
  if (!normalized || typeof (normalized as { toDate?: unknown }).toDate !== 'function') {
    return t('journal.form.datePlaceholder')
  }
  const jsDate = (normalized as { toDate: (zone: string) => Date }).toDate(getLocalTimeZone())
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(jsDate)
}

const calendarLabel = computed(() => formatCalendarLabel(calendarDate.value))

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

const updateEntryAmount = (index: number, value: string) => {
  const target = entries.value[index]
  if (!target) return
  target.amount_cents = formatCurrencyInput(value)
}

const handleCalendarChange = (value: unknown) => {
  calendarDate.value = value
  touchField('occurredAt')
}

const { touched, errors, touchField, validateAll } = useInlineValidation(
  { occurredAt, description },
  {
    occurredAt: dateSchema(t('journal.form.occurredAtLabel')),
    description: descriptionSchema(t('journal.form.descriptionLabel'))
  }
)

const accountMap = computed(() => new Map(accounts.value.map((item) => [item.id, item.name])))
const categoryMap = computed(() => new Map(categories.value.map((item) => [item.id, item.name])))

const currencyFormatter = computed(() => {
  const currency = ledgerContext.activeLedger.value?.currency_code || 'BRL'
  return new Intl.NumberFormat(locale.value, { style: 'currency', currency })
})

const formatAmount = (amount: number) => currencyFormatter.value.format(amount / 100)

const entryKindLabel = (kind: EntryKind) => t(`journal.kinds.${kind}`)

const accountFilterLabel = computed(() => {
  if (filterAccounts.value.length === 0) return t('journal.filters.allAccounts')
  if (filterAccounts.value.length === 1) {
    const accountId = filterAccounts.value[0]
    if (!accountId) return t('journal.filters.selectedCount', { count: 1 })
    return accountMap.value.get(accountId) || t('journal.filters.selectedCount', { count: 1 })
  }
  return t('journal.filters.selectedCount', { count: filterAccounts.value.length })
})

const categoryFilterLabel = computed(() => {
  if (filterCategories.value.length === 0) return t('journal.filters.allCategories')
  if (filterCategories.value.length === 1) {
    const categoryId = filterCategories.value[0]
    if (!categoryId) return t('journal.filters.selectedCount', { count: 1 })
    return categoryMap.value.get(categoryId) || t('journal.filters.selectedCount', { count: 1 })
  }
  return t('journal.filters.selectedCount', { count: filterCategories.value.length })
})


const journalRows = computed<JournalRow[]>(() => {
  const filteredAccounts = appliedFilters.value.account_ids
  const filteredCategories = appliedFilters.value.category_ids

  return transactions.value.flatMap((transaction) =>
    transaction.entries
      .filter((entry) => {
        if (filteredAccounts.length && !filteredAccounts.includes(entry.account_id)) return false
        if (filteredCategories.length && !filteredCategories.includes(entry.category_id || '')) return false
        return true
      })
      .map((entry) => ({
        row_id: `${transaction.id}-${entry.id}`,
        transaction_id: transaction.id,
        occurred_at: transaction.occurred_at,
        description: transaction.description,
        notes: transaction.notes,
        account_id: entry.account_id,
        category_id: entry.category_id,
        account_name: accountMap.value.get(entry.account_id) || entry.account_id,
        category_name: entry.category_id
          ? categoryMap.value.get(entry.category_id) || entry.category_id
          : t('journal.entry.noCategory'),
        kind: entry.kind,
        amount_cents: entry.amount_cents,
        memo: entry.memo
      }))
  )
})

const transactionById = computed(() => new Map(transactions.value.map((transaction) => [transaction.id, transaction])))

const columnLabels = computed(() => ({
  occurred_at: t('journal.columns.date'),
  description: t('journal.columns.description'),
  account: t('journal.columns.account'),
  category: t('journal.columns.category'),
  kind: t('journal.columns.kind'),
  amount: t('journal.columns.amount'),
  memo: t('journal.columns.memo'),
  actions: t('journal.columns.actions')
}))

const getColumnLabel = (id: string) =>
  columnLabels.value[id as keyof typeof columnLabels.value] || id


const columns: ColumnDef<JournalRow>[] = [
  {
    accessorKey: 'occurred_at',
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: columnLabels.value.occurred_at,
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) => h('div', { class: 'text-sm font-medium' }, new Date(row.original.occurred_at).toLocaleDateString(locale.value))
  },
  {
    accessorKey: 'description',
    header: () => h('div', { class: 'text-left' }, columnLabels.value.description),
    cell: ({ row }) =>
      h('div', { class: 'space-y-1' }, [
        h('div', { class: 'font-medium' }, row.original.description),
        row.original.notes ? h('div', { class: 'text-xs text-muted-foreground' }, row.original.notes) : null
      ])
  },
  {
    id: 'account',
    accessorFn: (row) => row.account_name,
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: columnLabels.value.account,
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) => h('div', { class: 'text-sm' }, row.original.account_name)
  },
  {
    id: 'category',
    accessorFn: (row) => row.category_name,
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: columnLabels.value.category,
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) => h('div', { class: 'text-sm text-muted-foreground' }, row.original.category_name)
  },
  {
    accessorKey: 'kind',
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: columnLabels.value.kind,
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) =>
      h('span', { class: 'rounded-full border border-border/60 px-2 py-0.5 text-[10px] uppercase text-muted-foreground' }, entryKindLabel(row.original.kind))
  },
  {
    accessorKey: 'amount_cents',
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: columnLabels.value.amount,
        align: 'right',
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) => h('div', { class: 'text-right font-medium' }, formatAmount(row.original.amount_cents))
  },
  {
    accessorKey: 'memo',
    header: () => h('div', { class: 'text-left' }, columnLabels.value.memo),
    cell: ({ row }) => h('div', { class: 'text-sm text-muted-foreground' }, row.original.memo || '')
  },
  {
    id: 'actions',
    enableHiding: false,
    header: () => h('div', { class: 'text-right' }, columnLabels.value.actions),
    cell: ({ row }) =>
      h(
        'div',
        { class: 'flex h-full items-center justify-end' },
        [
          h(
            DropdownMenu,
            null,
            {
              default: () => [
                h(
                  DropdownMenuTrigger,
                  { asChild: true },
                  {
                    default: () =>
                      h(
                        Button,
                        { variant: 'ghost', size: 'icon', class: 'h-8 w-8 p-0' },
                        { default: () => h(MoreHorizontal, { class: 'h-4 w-4' }) }
                      )
                  }
                ),
                h(
                  DropdownMenuContent,
                  { align: 'end' },
                  {
                    default: () => [
                      h(
                        DropdownMenuItem,
                        {
                          onClick: () => {
                            const transaction = transactionById.value.get(row.original.transaction_id)
                            if (transaction) openEdit(transaction)
                          }
                        },
                        { default: () => t('journal.actions.details') }
                      ),
                      h(
                        DropdownMenuItem,
                        {
                          onClick: () => {
                            const transaction = transactionById.value.get(row.original.transaction_id)
                            if (transaction) openEdit(transaction)
                          }
                        },
                        { default: () => t('journal.actions.edit') }
                      )
                    ]
                  }
                )
              ]
            }
          )
        ]
      )
  }
]

const sorting = ref<SortingState>([{ id: 'occurred_at', desc: true }])
const columnVisibility = ref<VisibilityState>({
  occurred_at: false,
  memo: false
})

const table = useVueTable({
  get data() {
    return journalRows.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
  getSortedRowModel: getSortedRowModel(),
  onSortingChange: (updaterOrValue) => valueUpdater(updaterOrValue, sorting),
  onColumnVisibilityChange: (updaterOrValue) => valueUpdater(updaterOrValue, columnVisibility),
  state: {
    get sorting() {
      return sorting.value
    },
    get columnVisibility() {
      return columnVisibility.value
    }
  }
})

const showLoadMore = computed(() => {
  const limit = Number(filterLimit.value) || 50
  return Boolean(nextCursor.value) && lastPageSize.value >= limit
})

const createEntryDraft = (): EntryDraft => ({
  id: globalThis.crypto?.randomUUID ? globalThis.crypto.randomUUID() : String(Date.now() + Math.random()),
  account_id: '',
  category_id: '__none__',
  kind: 'normal',
  amount_cents: '',
  memo: ''
})

const resetValidation = () => {
  touched.occurredAt = false
  touched.description = false
  errors.occurredAt = []
  errors.description = []
  entryErrors.value = []
}

const resetForm = () => {
  const today = new Date()
  occurredAt.value = today.toISOString().slice(0, 10)
  calendarDate.value = parseDate(occurredAt.value)
  description.value = ''
  notes.value = ''
  entries.value = [createEntryDraft()]
  formError.value = ''
  resetValidation()
}

const openCreate = () => {
  if (!canEdit.value) return
  formMode.value = 'create'
  editingTransaction.value = null
  resetForm()
  formOpen.value = true
}

const openEdit = (transaction: Transaction) => {
  if (!canEdit.value) return
  formMode.value = 'edit'
  editingTransaction.value = transaction
  occurredAt.value = transaction.occurred_at.slice(0, 10)
  calendarDate.value = parseDate(occurredAt.value)
  description.value = transaction.description
  notes.value = transaction.notes ?? ''
  entries.value = transaction.entries.map((entry) => ({
    id: entry.id,
    account_id: entry.account_id,
    category_id: entry.category_id ?? '__none__',
    kind: entry.kind,
    amount_cents: formatCurrencyInput(String(entry.amount_cents)),
    memo: entry.memo ?? ''
  }))
  formError.value = ''
  resetValidation()
  formOpen.value = true
}

const validateEntries = () => {
  entryErrors.value = entries.value.map((entry) => {
    const errorBag: EntryDraftError = {}
    if (!entry.account_id) {
      errorBag.account = t('journal.form.entryErrors.account')
    }
    if (!entry.kind) {
      errorBag.kind = t('journal.form.entryErrors.kind')
    }
    const amount = parseCurrencyToCents(entry.amount_cents)
    if (!amount || amount <= 0) {
      errorBag.amount = t('journal.form.entryErrors.amount')
    }
    return errorBag
  })
  return entryErrors.value.some((errorBag) => Object.keys(errorBag).length > 0)
}

const buildEntryPayload = (): TransactionEntryInput[] =>
  entries.value.map((entry) => ({
    account_id: entry.account_id,
    category_id: entry.category_id === '__none__' ? null : entry.category_id,
    kind: entry.kind,
    amount_cents: parseCurrencyToCents(entry.amount_cents),
    memo: entry.memo.trim() ? entry.memo.trim() : null
  }))

const handleSubmit = async () => {
  formError.value = ''
  const hasFieldErrors = validateAll()
  const hasEntryErrors = validateEntries()
  if (hasFieldErrors || hasEntryErrors) return
  isSaving.value = true
  try {
    const payload = {
      occurred_at: occurredAt.value,
      description: description.value.trim(),
      notes: notes.value.trim() ? notes.value.trim() : null,
      entries: buildEntryPayload()
    }
    if (formMode.value === 'create') {
      await createTransaction(payload)
      push.success({ title: t('journal.title'), message: t('journal.messages.created') })
    } else if (editingTransaction.value) {
      await updateTransaction(editingTransaction.value.id, payload)
      push.success({ title: t('journal.title'), message: t('journal.messages.updated') })
    }
    formOpen.value = false
  } catch (err) {
    if (isApiError(err)) {
      if (err.code === 'TRANSFER_NOT_BALANCED') {
        formError.value = t('journal.messages.unbalanced')
      } else if (err.code === 'TRANSACTION_REFERENCED') {
        formError.value = t('journal.messages.referenced')
      } else {
        formError.value = t('journal.messages.error')
      }
    } else {
      formError.value = t('journal.messages.error')
    }
  } finally {
    isSaving.value = false
  }
}


const removeEntry = (index: number) => {
  if (entries.value.length <= 1) return
  entries.value.splice(index, 1)
  entryErrors.value.splice(index, 1)
}

const addEntry = () => {
  entries.value.push(createEntryDraft())
  entryErrors.value.push({})
}

const applyFilters = async () => {
  const periodRange = journalPeriod.range.value
  const accountIdParam = filterAccounts.value.length === 1 ? filterAccounts.value[0] : ''
  const categoryIdParam = filterCategories.value.length === 1 ? filterCategories.value[0] : ''
  appliedFilters.value = {
    q: filterQuery.value.trim(),
    from: periodRange.from,
    to: periodRange.to,
    account_ids: [...filterAccounts.value],
    category_ids: [...filterCategories.value],
    limit: Number(filterLimit.value) || 50
  }
  await fetchTransactions({
    ...appliedFilters.value,
    account_id: accountIdParam || undefined,
    category_id: categoryIdParam || undefined
  })
}

const resetFilters = async () => {
  filterQuery.value = ''
  filterAccounts.value = []
  filterCategories.value = []
  filterLimit.value = '50'
  await applyFilters()
}

const loadMore = async () => {
  if (!nextCursor.value) return
  await fetchTransactions(
    {
      ...appliedFilters.value,
      cursor_occurred_at: nextCursor.value.cursor_occurred_at,
      cursor_id: nextCursor.value.cursor_id,
      account_id: appliedFilters.value.account_ids.length === 1 ? appliedFilters.value.account_ids[0] : undefined,
      category_id: appliedFilters.value.category_ids.length === 1 ? appliedFilters.value.category_ids[0] : undefined
    },
    { append: true }
  )
}

watch(
  () => ledgerContext.activeLedgerId.value,
  async (ledgerId) => {
    if (!ledgerId) return
    await Promise.all([fetchAccounts(), fetchCategories(), applyFilters()])
  },
  { immediate: true }
)

watch(
  () => calendarDate.value,
  (value) => {
    const normalized = normalizeCalendarValue(value)
    if (!normalized) return
    const iso = normalized.toString()
    if (occurredAt.value !== iso) {
      occurredAt.value = iso
    }
  }
)

watch(
  () => [journalPeriod.year.value, journalPeriod.month.value],
  async () => {
    if (!ledgerContext.activeLedgerId.value) return
    await applyFilters()
  }
)

watch(
  () => [filterAccounts.value, filterCategories.value, filterLimit.value],
  () => {
    if (!ledgerContext.activeLedgerId.value) return
    void applyFilters()
  }
)


watch(
  () => journalUi.createOpen.value,
  (open) => {
    if (open) {
      openCreate()
      journalUi.createOpen.value = false
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="grid gap-4">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-sm text-muted-foreground">{{ t('journal.description') }}</p>
      </div>

    </div>

    <CrudTableCard
      :title="t('journal.list.title')"
      :description="t('journal.list.description')"
      :loading="loading"
      :error="error"
      :empty="transactions.length === 0"
      :loading-message="t('journal.loading')"
      :empty-message="t('journal.empty')"
    >
      <template #toolbar>
        <div class="grid w-full gap-3 md:grid-cols-2 lg:grid-cols-4">
          <div>
            <Label class="text-xs text-muted-foreground">{{ t('journal.filters.query') }}</Label>
            <Input v-model="filterQuery" class="mt-1 h-9" :placeholder="t('journal.filters.queryPlaceholder')" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">{{ t('journal.filters.account') }}</Label>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="outline" class="mt-1 h-9 min-w-[180px] justify-between">
                  <span class="truncate text-left">{{ accountFilterLabel }}</span>
                  <ChevronDown class="ml-2 h-4 w-4 shrink-0" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" class="max-h-72 w-64 overflow-auto">
                <DropdownMenuItem @select.prevent="filterAccounts = []">
                  {{ t('journal.filters.clear') }}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuCheckboxItem
                  v-for="account in accounts"
                  :key="account.id"
                  :model-value="filterAccounts.includes(account.id)"
                  @select.prevent
                  @update:model-value="(value) => {
                    if (value) {
                      if (!filterAccounts.includes(account.id)) {
                        filterAccounts = [...filterAccounts, account.id]
                      }
                    } else {
                      filterAccounts = filterAccounts.filter((item) => item !== account.id)
                    }
                  }"
                >
                  {{ account.name }}
                </DropdownMenuCheckboxItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">{{ t('journal.filters.category') }}</Label>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="outline" class="mt-1 h-9 min-w-[180px] justify-between">
                  <span class="truncate text-left">{{ categoryFilterLabel }}</span>
                  <ChevronDown class="ml-2 h-4 w-4 shrink-0" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" class="max-h-72 w-64 overflow-auto">
                <DropdownMenuItem @select.prevent="filterCategories = []">
                  {{ t('journal.filters.clear') }}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuCheckboxItem
                  v-for="category in categories"
                  :key="category.id"
                  :model-value="filterCategories.includes(category.id)"
                  @select.prevent
                  @update:model-value="(value) => {
                    if (value) {
                      if (!filterCategories.includes(category.id)) {
                        filterCategories = [...filterCategories, category.id]
                      }
                    } else {
                      filterCategories = filterCategories.filter((item) => item !== category.id)
                    }
                  }"
                >
                  {{ category.name }}
                </DropdownMenuCheckboxItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">{{ t('journal.filters.limit') }}</Label>
            <Select v-model="filterLimit">
              <SelectTrigger class="mt-1 h-9 min-w-[120px]">
                <SelectValue :placeholder="t('journal.filters.limitPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="25">25</SelectItem>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
                <SelectItem value="200">200</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <div class="mt-3 flex flex-wrap items-center gap-2">
          <Button size="sm" @click="applyFilters">{{ t('journal.filters.apply') }}</Button>
          <Button size="sm" variant="outline" @click="resetFilters">{{ t('journal.filters.reset') }}</Button>
          <div class="ml-auto">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="outline" size="sm">
                  {{ t('journal.columns.toggle') }}
                  <ChevronDown class="ml-2 h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>{{ t('journal.columns.title') }}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuCheckboxItem
                  v-for="column in table.getAllColumns().filter((column) => column.getCanHide())"
                  :key="column.id"
                  :model-value="column.getIsVisible()"
                  @update:model-value="(value) => column.toggleVisibility(!!value)"
                >
                  {{ getColumnLabel(column.id) }}
                </DropdownMenuCheckboxItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </template>

      <div class="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
              <TableHead v-for="header in headerGroup.headers" :key="header.id">
                <FlexRender v-if="!header.isPlaceholder" :render="header.column.columnDef.header" :props="header.getContext()" />
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <template v-if="table.getRowModel().rows?.length">
              <TableRow v-for="row in table.getRowModel().rows" :key="row.id">
                <TableCell v-for="cell in row.getVisibleCells()" :key="cell.id">
                  <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
                </TableCell>
              </TableRow>
            </template>
            <TableRow v-else>
              <TableCell :colspan="table.getAllColumns().length" class="h-24 text-center">
                {{ t('journal.empty') }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <div v-if="showLoadMore" class="mt-4 flex justify-center">
        <Button variant="outline" size="sm" @click="loadMore">
          {{ t('journal.actions.loadMore') }}
          <ChevronDown class="ml-2 h-4 w-4" />
        </Button>
      </div>
    </CrudTableCard>

    <Sheet v-model:open="formOpen">
      <SheetContent side="right" class="w-full sm:max-w-2xl px-6 py-6">
        <SheetHeader>
          <SheetTitle>
            {{ formMode === 'create' ? t('journal.form.createTitle') : t('journal.form.editTitle') }}
          </SheetTitle>
          <SheetDescription>
            {{ formMode === 'create' ? t('journal.form.createDescription') : t('journal.form.editDescription') }}
          </SheetDescription>
        </SheetHeader>

        <div class="mt-6 grid gap-4">
          <div class="grid gap-3 md:grid-cols-3">
            <div class="grid gap-2">
              <Label for-id="journal_date">{{ t('journal.form.occurredAtLabel') }} <span class="text-destructive">*</span></Label>
              <Popover>
                <PopoverTrigger as-child>
                  <Button
                    id="journal_date"
                    variant="outline"
                    class="h-9 w-full justify-start gap-2 text-left font-normal md:w-[200px]"
                    :class="getInputClass({ touched: touched.occurredAt, hasError: Boolean(errors.occurredAt[0]) })"
                  >
                    <CalendarIcon class="h-4 w-4 text-muted-foreground" />
                    <span>{{ calendarLabel }}</span>
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0" align="start">
                  <Calendar :model-value="calendarDate as any" @update:model-value="handleCalendarChange" />
                </PopoverContent>
              </Popover>
              <p v-if="touched.occurredAt && errors.occurredAt[0]" class="text-xs text-destructive">
                {{ errors.occurredAt[0].message }}
              </p>
            </div>
            <div class="grid gap-2 md:col-span-2">
              <Label for-id="journal_notes">{{ t('journal.form.notesLabel') }}</Label>
              <Input id="journal_notes" v-model="notes" :placeholder="t('journal.form.notesPlaceholder')" />
            </div>
          </div>

          <div class="grid gap-2">
            <Label for-id="journal_description">{{ t('journal.form.descriptionLabel') }} <span class="text-destructive">*</span></Label>
            <Input
              id="journal_description"
              v-model="description"
              :class="getInputClass({ touched: touched.description, hasError: Boolean(errors.description[0]) })"
              @blur="touchField('description')"
            />
            <p v-if="touched.description && errors.description[0]" class="text-xs text-destructive">
              {{ errors.description[0].message }}
            </p>
          </div>

          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-sm font-medium">{{ t('journal.form.entriesTitle') }}</h3>
              <p class="text-xs text-muted-foreground">{{ t('journal.form.entriesDescription') }}</p>
            </div>
            <Button variant="outline" size="sm" type="button" @click="addEntry">
              <Plus class="mr-2 h-4 w-4" />
              {{ t('journal.form.addEntry') }}
            </Button>
          </div>

          <div class="grid gap-3">
            <div
              v-for="(entry, index) in entries"
              :key="entry.id"
              class="rounded-lg border border-border/60 p-3"
            >
              <div class="grid gap-3 md:grid-cols-6">
                <div class="md:col-span-2">
                  <Label class="text-xs text-muted-foreground">{{ t('journal.form.entryAccount') }}</Label>
                  <Select v-model="entry.account_id">
                    <SelectTrigger class="mt-1 h-9">
                      <SelectValue :placeholder="t('journal.form.entryAccountPlaceholder')" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="account in accounts" :key="account.id" :value="account.id">
                        {{ account.name }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <p v-if="entryErrors[index]?.account" class="text-xs text-destructive">
                    {{ entryErrors[index].account }}
                  </p>
                </div>
                <div class="md:col-span-2">
                  <Label class="text-xs text-muted-foreground">{{ t('journal.form.entryCategory') }}</Label>
                  <Select v-model="entry.category_id">
                    <SelectTrigger class="mt-1 h-9">
                      <SelectValue :placeholder="t('journal.form.entryCategoryPlaceholder')" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__none__">{{ t('journal.form.entryCategoryNone') }}</SelectItem>
                      <SelectItem v-for="category in categories" :key="category.id" :value="category.id">
                        {{ category.name }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label class="text-xs text-muted-foreground">{{ t('journal.form.entryKind') }}</Label>
                  <Select v-model="entry.kind">
                    <SelectTrigger class="mt-1 h-9">
                      <SelectValue :placeholder="t('journal.form.entryKindPlaceholder')" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="normal">{{ t('journal.kinds.normal') }}</SelectItem>
                      <SelectItem value="transfer">{{ t('journal.kinds.transfer') }}</SelectItem>
                      <SelectItem value="adjust">{{ t('journal.kinds.adjust') }}</SelectItem>
                    </SelectContent>
                  </Select>
                  <p v-if="entryErrors[index]?.kind" class="text-xs text-destructive">
                    {{ entryErrors[index].kind }}
                  </p>
                </div>
                <div class="md:col-span-6">
                  <div class="grid gap-3 md:grid-cols-[minmax(0,160px)_1fr_auto]">
                    <div class="grid gap-2">
                      <Label class="text-xs text-muted-foreground">{{ t('journal.form.entryAmount') }}</Label>
                      <Input
                        :model-value="entry.amount_cents"
                        class="h-9"
                        inputmode="decimal"
                        placeholder="00,00"
                        @update:model-value="(value) => updateEntryAmount(index, String(value))"
                      />
                      <p v-if="entryErrors[index]?.amount" class="text-xs text-destructive">
                        {{ entryErrors[index].amount }}
                      </p>
                    </div>
                    <div class="grid gap-2">
                      <Label class="text-xs text-muted-foreground">{{ t('journal.form.entryMemo') }}</Label>
                      <Input v-model="entry.memo" class="h-9" :placeholder="t('journal.form.entryMemoPlaceholder')" />
                    </div>
                    <div class="flex items-end">
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger as-child>
                            <span>
                              <Button
                                variant="ghost"
                                size="icon"
                                type="button"
                                :disabled="entries.length <= 1"
                                @click="removeEntry(index)"
                              >
                                <Trash2 class="h-4 w-4" />
                              </Button>
                            </span>
                          </TooltipTrigger>
                          <TooltipContent>
                            {{ t('journal.form.removeEntry') }}
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <p v-if="formError" class="text-sm text-destructive">
            {{ formError }}
          </p>
        </div>

        <SheetFooter class="mt-6 flex-row justify-end gap-2">
          <Button variant="outline" type="button" @click="formOpen = false">
            {{ t('journal.form.cancel') }}
          </Button>
          <Button type="button" :disabled="isSaving" @click="handleSubmit">
            {{ isSaving ? t('journal.form.saving') : t('journal.form.submit') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

  </div>
</template>
