<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import {
  type ColumnDef,
  type SortingState,
  type ColumnFiltersState,
  type VisibilityState,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useVueTable,
  FlexRender
} from '@tanstack/vue-table'
import { Button } from '@shared/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuCheckboxItem } from '@shared/components/ui/dropdown-menu'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@shared/components/ui/sheet'
import { Switch } from '@shared/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@shared/components/ui/table'
import { CornerDownRight, MoreHorizontal, ChevronDown } from 'lucide-vue-next'
import { push } from 'notivue'
import CrudTableCard from '@shared/components/CrudTableCard.vue'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'
import SortableColumnHeader from '@shared/components/SortableColumnHeader.vue'
import { useCategories, type Category, type CategoryDirection } from '#layers/categories/composables/useCategories'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { categoryDirectionSchema, nameSchema, useInlineValidation, getInputClass } from '@shared/validators'
import { isApiError } from '@shared/utils/api-error'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const ledgerContext = useLedgerContext()
const canEdit = computed(() => ledgerContext.hasRole('editor'))

const {
  categories,
  loading,
  error,
  fetchCategories,
  createCategory,
  updateCategory,
  deactivateCategory,
  seedCategories
} = useCategories()

// --- Data Preparation (Hierarchy) ---

type CategoryRow = Category & { depth: number }

const buildHierarchy = (items: Category[]) => {
  const sorted = [...items].sort((a, b) => a.name.localeCompare(b.name))
  const byParent = new Map<string | null, Category[]>()
  for (const item of sorted) {
    const key = item.parent_id ?? null
    const list = byParent.get(key) ?? []
    list.push(item)
    byParent.set(key, list)
  }

  const walk = (parentKey: string | null, depth: number, acc: CategoryRow[]) => {
    const children = byParent.get(parentKey) ?? []
    for (const child of children) {
      acc.push({ ...child, depth })
      walk(child.id, depth + 1, acc)
    }
  }

  const output: CategoryRow[] = []
  walk(null, 0, output)
  return output
}

const flattenedCategories = computed<CategoryRow[]>(() => buildHierarchy(categories.value))

type CategorySuggestion = {
  name: string
  direction: CategoryDirection
  isBudgetRelevant: boolean
  isBudgetBase: boolean
}

const suggestedCategories: CategorySuggestion[] = [
  { name: 'Gastos fixos', direction: 'out', isBudgetRelevant: true, isBudgetBase: false },
  { name: 'Conforto', direction: 'out', isBudgetRelevant: true, isBudgetBase: false },
  { name: 'Lazer', direction: 'out', isBudgetRelevant: true, isBudgetBase: false },
  { name: 'Investimentos (Saída)', direction: 'out', isBudgetRelevant: true, isBudgetBase: false },
  { name: 'Objetivos', direction: 'out', isBudgetRelevant: true, isBudgetBase: false },
  { name: 'Educação', direction: 'out', isBudgetRelevant: true, isBudgetBase: false },
  { name: 'Investimentos (Entrada)', direction: 'in', isBudgetRelevant: false, isBudgetBase: false },
  { name: 'Salário', direction: 'in', isBudgetRelevant: false, isBudgetBase: true },
  { name: 'Extra', direction: 'in', isBudgetRelevant: false, isBudgetBase: true },
  { name: 'Terceiros', direction: 'in', isBudgetRelevant: false, isBudgetBase: true }
]

const normalizeCategoryName = (value: string) => value.trim().toLowerCase()

const existingCategoryNames = computed(
  () => new Set(categories.value.map((item) => normalizeCategoryName(item.name)))
)

const missingSuggestions = computed(() =>
  suggestedCategories.filter(
    (suggestion) => !existingCategoryNames.value.has(normalizeCategoryName(suggestion.name))
  )
)

const missingOutSuggestions = computed(() =>
  missingSuggestions.value.filter((item) => item.direction === 'out')
)

const missingInSuggestions = computed(() =>
  missingSuggestions.value.filter((item) => item.direction === 'in')
)

const hasSuggestions = computed(() => missingSuggestions.value.length > 0)

const seedSelection = ref<string[]>([])
const seedRelevance = ref<Record<string, boolean>>({})
const seedBase = ref<Record<string, boolean>>({})
const seedSaving = ref(false)

const getSeedRelevance = (name: string, fallback: boolean) =>
  seedRelevance.value[name] ?? fallback

const getSeedBase = (name: string, fallback: boolean) =>
  seedBase.value[name] ?? fallback

const setSeedRelevance = (name: string, value: boolean) => {
  seedRelevance.value = { ...seedRelevance.value, [name]: value }
}

const setSeedBase = (name: string, value: boolean) => {
  seedBase.value = { ...seedBase.value, [name]: value }
}

const resetSeedSelection = () => {
  seedSelection.value = missingSuggestions.value.map((item) => item.name)
  seedRelevance.value = missingSuggestions.value.reduce<Record<string, boolean>>((acc, item) => {
    acc[item.name] = item.isBudgetRelevant
    return acc
  }, {})
  seedBase.value = missingSuggestions.value.reduce<Record<string, boolean>>((acc, item) => {
    acc[item.name] = item.isBudgetBase
    return acc
  }, {})
}

const toggleSeedSelection = (name: string, enabled: boolean) => {
  if (enabled) {
    if (!seedSelection.value.includes(name)) {
      seedSelection.value = [...seedSelection.value, name]
    }
    return
  }
  seedSelection.value = seedSelection.value.filter((item) => item !== name)
}

const showSeedInline = computed(() => categories.value.length === 0 && hasSuggestions.value)

watch(
  () => missingSuggestions.value,
  () => {
    if (missingSuggestions.value.length === 0) {
      seedSelection.value = []
      seedRelevance.value = {}
      seedBase.value = {}
      return
    }
    resetSeedSelection()
  },
  { immediate: true }
)

// --- TanStack Table Configuration ---

const sorting = ref<SortingState>([])
const columnFilters = ref<ColumnFiltersState>([])
const globalFilter = ref('')
const columnVisibility = ref<VisibilityState>({})

const directionLabel = (value: string) =>
  value === 'in' ? t('categories.directions.in') : t('categories.directions.out')

const statusLabel = (value: boolean) =>
  value ? t('categories.status.active') : t('categories.status.inactive')

const columns: ColumnDef<CategoryRow>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: t('categories.table.name'),
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) => {
       const depth = row.original.depth
       return h('div', { class: 'flex items-center gap-2', style: { paddingLeft: `${depth * 12}px` } }, [
          depth > 0 ? h(CornerDownRight, { class: 'h-3.5 w-3.5 text-muted-foreground' }) : null,
          h('div', {}, [
             h('div', { class: 'font-medium text-foreground' }, row.original.name),
           //  h('div', { class: 'text-xs text-muted-foreground' }, row.original.id) // ID is noisy, hiding
          ])
       ])
    }
  },
  {
    accessorKey: 'direction',
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: t('categories.table.direction'),
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) => h('span', { class: 'text-muted-foreground' }, directionLabel(row.original.direction))
  },
  {
    id: 'flags',
    header: t('categories.table.flags'),
    cell: ({ row }) => {
       const children = []
       if (row.original.is_budget_base) {
         children.push(h('span', { class: 'inline-flex items-center rounded-full border border-primary/30 bg-primary/10 px-2 py-1 text-xs text-primary mr-1' }, t('categories.flags.budgetBase')))
       }
       if (row.original.is_budget_relevant) {
         children.push(h('span', { class: 'inline-flex items-center rounded-full border border-emerald-500/30 bg-emerald-500/10 px-2 py-1 text-xs text-emerald-600 dark:text-emerald-300 mr-1' }, t('categories.flags.budgetRelevant')))
       }
       if (children.length === 0) {
          children.push(h('span', { class: 'text-xs text-muted-foreground' }, t('categories.flags.none')))
       }
       return h('div', { class: 'flex flex-wrap' }, children)
    }
  },
  {
    accessorKey: 'is_active',
    header: ({ column }) =>
      h(SortableColumnHeader, {
        label: t('categories.table.status'),
        onToggle: () => column.toggleSorting(column.getIsSorted() === 'asc')
      }),
    cell: ({ row }) => h('span', {
       class: `inline-flex items-center rounded-full border px-2 py-1 text-xs ${row.original.is_active
          ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
          : 'border-muted-foreground/30 bg-muted/50 text-muted-foreground'}`
    }, statusLabel(row.original.is_active))
  },
  {
    id: 'actions',
    enableHiding: false,
    header: () => h('div', { class: 'text-right' }, t('categories.table.actions')),
    cell: ({ row }) => {
      const category = row.original
      return h(DropdownMenu, {}, { default: () => [
         h(DropdownMenuTrigger, { asChild: true }, { default: () => 
            h(Button, { variant: 'ghost', size: 'icon', disabled: !canEdit.value }, { default: () => h(MoreHorizontal, { class: 'h-4 w-4' }) })
         }),
         h(DropdownMenuContent, { align: 'end' }, { default: () => [
            h(DropdownMenuItem, { onClick: () => openEdit(category) }, { default: () => t('categories.actions.edit') }),
            h(DropdownMenuItem, { onClick: () => handleToggleActive(category) }, { 
                default: () => category.is_active ? t('categories.actions.deactivate') : t('categories.actions.activate') 
            })
         ]})
      ]})
    }
  }
]

const table = useVueTable({
  get data() { return flattenedCategories.value },
  columns,
  getCoreRowModel: getCoreRowModel(),
  getSortedRowModel: getSortedRowModel(),
  getFilteredRowModel: getFilteredRowModel(),
  state: {
    get sorting() { return sorting.value },
    get columnFilters() { return columnFilters.value },
    get globalFilter() { return globalFilter.value },
    get columnVisibility() { return columnVisibility.value },
  },
  onSortingChange: updater => {
      if (typeof updater === 'function') sorting.value = updater(sorting.value)
      else sorting.value = updater
  },
  onColumnFiltersChange: updater => {
      if (typeof updater === 'function') columnFilters.value = updater(columnFilters.value)
      else columnFilters.value = updater
  },
  onGlobalFilterChange: updater => {
      if (typeof updater === 'function') globalFilter.value = updater(globalFilter.value)
      else globalFilter.value = updater
  },
  onColumnVisibilityChange: updater => {
       if (typeof updater === 'function') columnVisibility.value = updater(columnVisibility.value)
       else columnVisibility.value = updater
  },
  globalFilterFn: (row, columnId, filterValue) => {
      const search = filterValue.toLowerCase()
      const name = row.original.name.toLowerCase()
      return name.includes(search)
  }
})


// --- State for Forms ---

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editingCategory = ref<Category | null>(null)
const isSaving = ref(false)
const formError = ref('')
const confirmOpen = ref(false)
const confirmCategory = ref<Category | null>(null)

const name = ref('')
const direction = ref<CategoryDirection>('out')
const parentId = ref<string | null>('__none__')
const isBudgetBase = ref(false)
const isBudgetRelevant = ref(true)

const { touched, errors, touchField, validateAll } = useInlineValidation(
  { name, direction },
  { name: nameSchema(t('categories.form.nameLabel')), direction: categoryDirectionSchema(t('categories.form.directionLabel')) }
)

const resetValidation = () => {
  touched.name = false
  touched.direction = false
  errors.name = []
  errors.direction = []
}

const handleSeed = async () => {
  if (!canEdit.value || seedSelection.value.length === 0) return
  seedSaving.value = true
  try {
    const selectedSuggestions = missingSuggestions.value.filter((item) =>
      seedSelection.value.includes(item.name)
    )
    const result = await seedCategories({
      preset: 'default_v1',
      items: selectedSuggestions.map((item) => ({
        name: item.name,
        direction: item.direction,
        is_budget_relevant:
          item.direction === 'out' ? getSeedRelevance(item.name, item.isBudgetRelevant) : undefined,
        is_budget_base: item.direction === 'in' ? getSeedBase(item.name, item.isBudgetBase) : undefined
      }))
    })
    await fetchCategories()
    const createdCount = result?.created?.length ?? 0
    push.success({
      title: t('categories.seed.title'),
      message: t('categories.seed.messages.added', { count: createdCount })
    })
  } catch (err) {
    if (isApiError(err) && err.code === 'VALIDATION_ERROR') {
      push.error({ title: t('categories.seed.title'), message: t('categories.seed.messages.invalid') })
    } else {
      push.error({ title: t('categories.seed.title'), message: t('categories.seed.messages.error') })
    }
  } finally {
    seedSaving.value = false
  }
}

const openEdit = (category: Category) => {
  formMode.value = 'edit'
  editingCategory.value = category
  name.value = category.name
  direction.value = category.direction
  parentId.value = category.parent_id ?? '__none__'
  isBudgetBase.value = category.is_budget_base
  isBudgetRelevant.value = category.is_budget_relevant
  formError.value = ''
  resetValidation()
  formOpen.value = true
}

const requestDeactivate = (category: Category) => {
  confirmCategory.value = category
  confirmOpen.value = true
}

const confirmDeactivate = async () => {
  if (!confirmCategory.value) return
  try {
    await deactivateCategory(confirmCategory.value.id)
    push.success({ title: t('categories.title'), message: t('categories.messages.deactivated') })
  } catch (err) {
    if (isApiError(err) && err.code === 'VALIDATION_ERROR') {
      push.error({ title: t('categories.title'), message: t('categories.messages.inUse') })
    } else {
      push.error({ title: t('categories.title'), message: t('categories.messages.error') })
    }
  } finally {
    confirmCategory.value = null
  }
}

const handleSubmit = async () => {
  formError.value = ''
  if (validateAll()) {
    return
  }
  isSaving.value = true
  try {
    if (formMode.value === 'create') {
      await createCategory({
        name: name.value,
        direction: direction.value,
        parent_id: parentId.value === '__none__' ? undefined : parentId.value,
        is_budget_base: isBudgetBase.value,
        is_budget_relevant: isBudgetRelevant.value
      })
      push.success({ title: t('categories.title'), message: t('categories.messages.created') })
    } else if (editingCategory.value) {
      await updateCategory(editingCategory.value.id, {
        name: name.value,
        parent_id: parentId.value === '__none__' ? undefined : parentId.value,
        is_budget_base: isBudgetBase.value,
        is_budget_relevant: isBudgetRelevant.value,
        is_active: editingCategory.value.is_active
      })
      push.success({ title: t('categories.title'), message: t('categories.messages.updated') })
    }
    formOpen.value = false
  } catch (err) {
    if (isApiError(err) && err.code === 'CONFLICT_DUPLICATE_NAME') {
      formError.value = t('categories.messages.duplicateName')
    } else {
      formError.value = t('categories.messages.error')
    }
  } finally {
    isSaving.value = false
  }
}

const handleToggleActive = async (category: Category) => {
  if (!canEdit.value) return
  try {
    if (category.is_active) {
      requestDeactivate(category)
      return
    }
    await updateCategory(category.id, {
      name: category.name,
      parent_id: category.parent_id ?? undefined,
      is_budget_base: category.is_budget_base,
      is_budget_relevant: category.is_budget_relevant,
      is_active: true
    })
    push.success({ title: t('categories.title'), message: t('categories.messages.activated') })
  } catch {
    push.error({ title: t('categories.title'), message: t('categories.messages.error') })
  }
}

const directionFlags = computed(() => {
  const isIn = direction.value === 'in'
  const isOut = direction.value === 'out'
  return {
    budgetBaseDisabled: isOut,
    budgetRelevantDisabled: isIn
  }
})

watch(
  () => direction.value,
  (value) => {
    if (value === 'in') {
      isBudgetRelevant.value = false
    }
    if (value === 'out') {
      isBudgetBase.value = false
    }
  }
)

const parentOptions = computed(() =>
  buildHierarchy(categories.value)
    .filter((item) => item.id !== editingCategory.value?.id)
    .map((item) => ({
      id: item.id,
      name: item.name,
      depth: item.depth
    }))
)

watch(
  () => ledgerContext.activeLedgerId.value,
  async (ledgerId) => {
    if (!ledgerId) return
    seedSelection.value = []
    seedRelevance.value = {}
    await fetchCategories()
  },
  { immediate: true }
)

// Reset filters to default state (showing everything or filtered?)
// Journal uses server-side filtering. Here we use client-side.
// We can expose filter controls bound to table state.
</script>

<template>
  <div class="grid gap-6">
    <CrudTableCard
      :title="t('categories.list.title')"
      :description="t('categories.list.description')"
      :loading="loading"
      :error="error"
      :empty="categories.length === 0"
      :loading-message="t('categories.loading')"
      :empty-message="t('categories.empty')"
    >
      <template #empty>
        <div v-if="showSeedInline" class="grid gap-4">
          <div>
            <p class="text-sm font-medium text-foreground">{{ t('categories.seed.title') }}</p>
            <p class="text-xs text-muted-foreground">{{ t('categories.seed.description') }}</p>
          </div>
          <div class="grid gap-4">
            <div v-if="missingInSuggestions.length" class="grid gap-2">
              <div class="flex items-center gap-3">
                <div class="h-px flex-1 bg-border/60"></div>
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {{ t('categories.seed.groups.in') }}
                </p>
                <div class="h-px flex-1 bg-border/60"></div>
              </div>
              <div class="grid gap-2">
                <div
                  v-for="suggestion in missingInSuggestions"
                  :key="suggestion.name"
                  class="grid gap-2 rounded-md border border-border/60 px-3 py-2"
                >
                  <div class="flex items-center justify-between gap-3">
                    <p class="text-sm font-medium">{{ suggestion.name }}</p>
                    <Switch
                      :disabled="!canEdit || seedSaving"
                      :checked="seedSelection.includes(suggestion.name)"
                      @update:checked="(value: boolean) => toggleSeedSelection(suggestion.name, value)"
                    />
                  </div>
                  <div class="flex items-center justify-between text-xs text-muted-foreground">
                    <span>
                      {{ suggestion.direction === 'in' ? t('categories.seed.budgetBase') : t('categories.seed.budgetRelevant') }}
                    </span>
                    <Switch
                      :disabled="!canEdit || seedSaving"
                      :checked="suggestion.direction === 'in'
                        ? getSeedBase(suggestion.name, suggestion.isBudgetBase)
                        : getSeedRelevance(suggestion.name, suggestion.isBudgetRelevant)"
                      @update:checked="(value: boolean) =>
                        suggestion.direction === 'in'
                          ? setSeedBase(suggestion.name, value)
                          : setSeedRelevance(suggestion.name, value)"
                    />
                  </div>
                </div>
              </div>
            </div>
            <div v-if="missingOutSuggestions.length" class="grid gap-2">
              <div class="flex items-center gap-3">
                <div class="h-px flex-1 bg-border/60"></div>
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {{ t('categories.seed.groups.out') }}
                </p>
                <div class="h-px flex-1 bg-border/60"></div>
              </div>
              <div class="grid gap-2">
                <div
                  v-for="suggestion in missingOutSuggestions"
                  :key="suggestion.name"
                  class="grid gap-2 rounded-md border border-border/60 px-3 py-2"
                >
                  <div class="flex items-center justify-between gap-3">
                    <p class="text-sm font-medium">{{ suggestion.name }}</p>
                    <Switch
                      :disabled="!canEdit || seedSaving"
                      :checked="seedSelection.includes(suggestion.name)"
                      @update:checked="(value: boolean) => toggleSeedSelection(suggestion.name, value)"
                    />
                  </div>
                  <div class="flex items-center justify-between text-xs text-muted-foreground">
                    <span>
                      {{ suggestion.direction === 'in' ? t('categories.seed.budgetBase') : t('categories.seed.budgetRelevant') }}
                    </span>
                    <Switch
                      :disabled="!canEdit || seedSaving"
                      :checked="suggestion.direction === 'in'
                        ? getSeedBase(suggestion.name, suggestion.isBudgetBase)
                        : getSeedRelevance(suggestion.name, suggestion.isBudgetRelevant)"
                      @update:checked="(value: boolean) =>
                        suggestion.direction === 'in'
                          ? setSeedBase(suggestion.name, value)
                          : setSeedRelevance(suggestion.name, value)"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button
              :disabled="!canEdit || seedSaving || seedSelection.length === 0"
              @click="handleSeed()"
            >
              {{ seedSaving ? t('categories.seed.saving') : t('categories.seed.submit') }}
            </Button>
          </div>
        </div>
        <div v-else class="flex flex-col items-center gap-3 text-center">
          <p class="text-sm text-muted-foreground">{{ t('categories.empty') }}</p>
        </div>
      </template>

      <template #toolbar>
         <div class="flex w-full items-center justify-between gap-4 flex-wrap">
             <div class="flex flex-1 items-center gap-2 max-w-sm">
                <Input 
                   :placeholder="t('categories.filters.placeholder')" 
                   :model-value="globalFilter"
                   @update:model-value="globalFilter = String($event)"
                   class="h-8 w-full"
                />
             </div>
             <div class="flex items-center gap-2">
                 <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                        <Button variant="outline" size="sm" class="ml-auto">
                            {{ t('journal.columns.toggle') || 'Columns' }} <ChevronDown class="ml-2 h-4 w-4" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuCheckboxItem
                            v-for="column in table.getAllColumns().filter(c => c.getCanHide())"
                            :key="column.id"
                            :checked="column.getIsVisible()"
                            @update:checked="(value: boolean) => column.toggleVisibility(!!value)"
                        >
                            {{ column.columnDef.header }}
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
                  <TableCell :colspan="columns.length" class="h-24 text-center">
                     {{ t('categories.empty') }}
                  </TableCell>
               </TableRow>
            </TableBody>
         </Table>
      </div>
    </CrudTableCard>

    <Sheet v-model:open="formOpen">
      <SheetContent side="right" class="w-full sm:max-w-md px-6 py-6">
        <SheetHeader>
          <SheetTitle>
            {{ formMode === 'create' ? t('categories.form.createTitle') : t('categories.form.editTitle') }}
          </SheetTitle>
          <SheetDescription>
            {{ formMode === 'create' ? t('categories.form.createDescription') : t('categories.form.editDescription') }}
          </SheetDescription>
        </SheetHeader>

        <div class="mt-6 grid gap-4">
          <div class="grid gap-2">
            <Label for-id="category_name">{{ t('categories.form.nameLabel') }} <span class="text-destructive">*</span></Label>
            <Input
              id="category_name"
              v-model="name"
              :class="getInputClass({ touched: touched.name, hasError: Boolean(errors.name[0]) })"
              @blur="touchField('name')"
            />
            <p v-if="touched.name && errors.name[0]" class="text-xs text-destructive">
              {{ errors.name[0].message }}
            </p>
          </div>
          <div class="grid gap-2">
            <Label>{{ t('categories.form.directionLabel') }} <span class="text-destructive">*</span></Label>
            <Select v-model="direction" :disabled="formMode === 'edit'">
              <SelectTrigger
                :class="getInputClass({ touched: touched.direction, hasError: Boolean(errors.direction[0]) })"
                @blur="touchField('direction')"
              >
                <SelectValue :placeholder="t('categories.form.directionPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="out">{{ t('categories.directions.out') }}</SelectItem>
                <SelectItem value="in">{{ t('categories.directions.in') }}</SelectItem>
              </SelectContent>
            </Select>
            <p v-if="touched.direction && errors.direction[0]" class="text-xs text-destructive">
              {{ errors.direction[0].message }}
            </p>
          </div>
          <div class="grid gap-2">
            <Label>{{ t('categories.form.parentLabel') }}</Label>
            <Select v-model="parentId">
              <SelectTrigger>
                <SelectValue :placeholder="t('categories.form.parentPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">{{ t('categories.form.parentNone') }}</SelectItem>
                <SelectItem v-for="option in parentOptions" :key="option.id" :value="option.id">
                  <span class="flex items-center gap-2" :style="{ paddingLeft: `${option.depth * 12}px` }">
                    <CornerDownRight v-if="option.depth > 0" class="h-3.5 w-3.5 text-muted-foreground" />
                    <span>{{ option.name }}</span>
                  </span>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="grid gap-3 rounded-lg border border-border/60 p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium">{{ t('categories.form.budgetBaseLabel') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('categories.form.budgetBaseHint') }}</p>
              </div>
              <Switch v-model:checked="isBudgetBase" :disabled="directionFlags.budgetBaseDisabled" />
            </div>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium">{{ t('categories.form.budgetRelevantLabel') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('categories.form.budgetRelevantHint') }}</p>
              </div>
              <Switch v-model:checked="isBudgetRelevant" :disabled="directionFlags.budgetRelevantDisabled" />
            </div>
            <p v-if="directionFlags.budgetBaseDisabled" class="text-xs text-muted-foreground">
              {{ t('categories.form.budgetBaseDisabledHint') }}
            </p>
            <p v-if="directionFlags.budgetRelevantDisabled" class="text-xs text-muted-foreground">
              {{ t('categories.form.budgetRelevantDisabledHint') }}
            </p>
          </div>
          <p v-if="formError" class="text-sm text-destructive">
            {{ formError }}
          </p>
        </div>

        <SheetFooter class="mt-6 flex-row justify-end gap-2">
          <Button variant="outline" type="button" @click="formOpen = false">
            {{ t('categories.form.cancel') }}
          </Button>
          <Button type="button" :disabled="isSaving" @click="handleSubmit">
            {{ isSaving ? t('categories.form.saving') : t('categories.form.submit') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="t('categories.confirm.deactivateTitle')"
      :description="confirmCategory ? t('categories.confirm.deactivateMessage', { name: confirmCategory.name }) : ''"
      :confirm-label="t('categories.confirm.confirmAction')"
      :cancel-label="t('categories.confirm.cancelAction')"
      @confirm="confirmDeactivate"
      @cancel="confirmOpen = false"
    />
  </div>
</template>
