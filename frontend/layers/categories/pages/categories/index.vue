<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Button } from '@shared/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@shared/components/ui/dropdown-menu'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@shared/components/ui/sheet'
import { Switch } from '@shared/components/ui/switch'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@shared/components/ui/tooltip'
import { CornerDownRight, MoreHorizontal, Plus } from 'lucide-vue-next'
import { push } from 'notivue'
import CrudTableCard from '@shared/components/CrudTableCard.vue'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'
import { useCategories, type Category, type CategoryDirection } from '#layers/categories/composables/useCategories'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { categoryDirectionSchema, nameSchema, useInlineValidation, getInputClass } from '@shared/validators'
import { isApiError } from '@shared/utils/api-error'

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
  deactivateCategory
} = useCategories()

const statusFilter = ref<'all' | 'active' | 'inactive'>('active')
const directionFilter = ref<'all' | CategoryDirection>('all')

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

const directionLabel = (value: CategoryDirection) =>
  value === 'in' ? t('categories.directions.in') : t('categories.directions.out')

const statusLabel = (value: boolean) =>
  value ? t('categories.status.active') : t('categories.status.inactive')

const resetValidation = () => {
  touched.name = false
  touched.direction = false
  errors.name = []
  errors.direction = []
}

const resetForm = () => {
  name.value = ''
  direction.value = 'out'
  parentId.value = '__none__'
  isBudgetBase.value = false
  isBudgetRelevant.value = true
  formError.value = ''
  resetValidation()
}

const openCreate = () => {
  formMode.value = 'create'
  editingCategory.value = null
  resetForm()
  formOpen.value = true
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

type CategoryRow = Category & { depth: number }

const filteredCategories = computed(() => {
  return categories.value.filter((item) => {
    const matchesDirection =
      directionFilter.value === 'all' || item.direction === directionFilter.value
    const matchesStatus =
      statusFilter.value === 'all' ||
      (statusFilter.value === 'active' ? item.is_active : !item.is_active)
    return matchesDirection && matchesStatus
  })
})

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

const flattenedCategories = computed<CategoryRow[]>(() => buildHierarchy(filteredCategories.value))

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
  () => [ledgerContext.activeLedgerId.value, statusFilter.value, directionFilter.value],
  async ([ledgerId]) => {
    if (!ledgerId) return
    await fetchCategories()
  },
  { immediate: true }
)
</script>

<template>
  <div class="grid gap-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold">{{ t('categories.title') }}</h1>
        <p class="text-sm text-muted-foreground">{{ t('categories.description') }}</p>
      </div>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger as-child>
            <span>
              <Button :disabled="!canEdit" @click="openCreate">
                <Plus class="h-4 w-4" />
                {{ t('categories.actions.new') }}
              </Button>
            </span>
          </TooltipTrigger>
          <TooltipContent v-if="!canEdit">
            {{ t('categories.readOnlyHint') }}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>

    <CrudTableCard
      :title="t('categories.list.title')"
      :description="t('categories.list.description')"
      :loading="loading"
      :error="error"
      :empty="flattenedCategories.length === 0"
      :loading-message="t('categories.loading')"
      :empty-message="t('categories.empty')"
    >
      <template #toolbar>
        <div class="grid w-full gap-3 md:w-[420px] md:grid-cols-2">
          <div>
            <Label class="text-xs text-muted-foreground">{{ t('categories.filters.direction') }}</Label>
            <Select v-model="directionFilter">
              <SelectTrigger class="mt-2 w-full">
                <SelectValue :placeholder="t('categories.filters.placeholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{{ t('categories.filters.allDirections') }}</SelectItem>
                <SelectItem value="out">{{ t('categories.directions.out') }}</SelectItem>
                <SelectItem value="in">{{ t('categories.directions.in') }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">{{ t('categories.filters.status') }}</Label>
            <Select v-model="statusFilter">
              <SelectTrigger class="mt-2 w-full">
                <SelectValue :placeholder="t('categories.filters.placeholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{{ t('categories.filters.all') }}</SelectItem>
                <SelectItem value="active">{{ t('categories.filters.active') }}</SelectItem>
                <SelectItem value="inactive">{{ t('categories.filters.inactive') }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </template>

      <div class="overflow-hidden rounded-lg border">
        <table class="w-full text-sm">
          <thead class="bg-muted/50 text-xs uppercase text-muted-foreground">
            <tr>
              <th class="px-4 py-3 text-left">{{ t('categories.table.name') }}</th>
              <th class="px-4 py-3 text-left">{{ t('categories.table.direction') }}</th>
              <th class="px-4 py-3 text-left">{{ t('categories.table.flags') }}</th>
              <th class="px-4 py-3 text-left">{{ t('categories.table.status') }}</th>
              <th class="px-4 py-3 text-right">{{ t('categories.table.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="category in flattenedCategories" :key="category.id" class="border-t">
              <td class="px-4 py-3">
                <div class="flex items-center gap-2" :style="{ paddingLeft: `${category.depth * 12}px` }">
                  <CornerDownRight v-if="category.depth > 0" class="h-3.5 w-3.5 text-muted-foreground" />
                  <div>
                    <div class="font-medium text-foreground">{{ category.name }}</div>
                    <div class="text-xs text-muted-foreground">{{ category.id }}</div>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3 text-muted-foreground">
                {{ directionLabel(category.direction) }}
              </td>
              <td class="px-4 py-3 text-muted-foreground">
                <div class="flex flex-wrap gap-2">
                  <span
                    v-if="category.is_budget_base"
                    class="inline-flex items-center rounded-full border border-primary/30 bg-primary/10 px-2 py-1 text-xs text-primary"
                  >
                    {{ t('categories.flags.budgetBase') }}
                  </span>
                  <span
                    v-if="category.is_budget_relevant"
                    class="inline-flex items-center rounded-full border border-emerald-500/30 bg-emerald-500/10 px-2 py-1 text-xs text-emerald-600 dark:text-emerald-300"
                  >
                    {{ t('categories.flags.budgetRelevant') }}
                  </span>
                  <span v-if="!category.is_budget_base && !category.is_budget_relevant" class="text-xs text-muted-foreground">
                    {{ t('categories.flags.none') }}
                  </span>
                </div>
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center rounded-full border px-2 py-1 text-xs"
                  :class="category.is_active
                    ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                    : 'border-muted-foreground/30 bg-muted/50 text-muted-foreground'"
                >
                  {{ statusLabel(category.is_active) }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" :disabled="!canEdit">
                      <MoreHorizontal class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="openEdit(category)">
                      {{ t('categories.actions.edit') }}
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="handleToggleActive(category)">
                      {{ category.is_active ? t('categories.actions.deactivate') : t('categories.actions.activate') }}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </td>
            </tr>
          </tbody>
        </table>
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
