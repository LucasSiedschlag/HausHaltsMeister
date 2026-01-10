<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@shared/components/ui/button'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@shared/components/ui/sheet'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@shared/components/ui/table'
import { Badge } from '@shared/components/ui/badge'
import { Plus, MoreHorizontal } from 'lucide-vue-next'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@shared/components/ui/dropdown-menu'
import { push } from 'notivue'
import CrudTableCard from '@shared/components/CrudTableCard.vue'
import { useBudget, type BudgetVersion } from '../../composables/useBudget'
import { useCategories } from '#layers/categories/composables/useCategories'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { isApiError } from '@shared/utils/api-error'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

const { t, locale, te } = useI18n()
const router = useRouter()
const { versions, loading, error, fetchVersions, createVersion, deleteVersion } = useBudget()
const { categories, fetchCategories } = useCategories()
const { setHeaderAction } = useHeaderAction()

const formOpen = ref(false)
const isSaving = ref(false)
const effectiveMonth = ref('')

// Local state for lines in the create form
const initialLines = ref<{ categoryId: string; name: string; percent: number }[]>([])

// Initialize default month to next month and prepopulate lines
const initForm = async () => {
  const next = new Date()
  next.setMonth(next.getMonth() + 1)
  effectiveMonth.value = next.toISOString().slice(0, 7) // YYYY-MM
  
  // Ensure we have fresh categories
  await fetchCategories({ active: true })

  // Filter relevant categories for initial population
  // Backend requires direction=out and IsBudgetRelevant=true
  initialLines.value = categories.value
    .filter(c => c.direction.toLowerCase() === 'out' && c.is_budget_relevant)
    .map(c => ({
      categoryId: c.id,
      name: c.name,
      percent: 0
    }))
}

const openCreate = async () => {
  try {
      await initForm()
      formOpen.value = true
  } catch {
      push.error(t('budget.messages.error'))
  }
}

setHeaderAction(
  { key: 'budget:new-version', labelKey: 'budget.actions.newVersion', requiresEditor: true },
  () => {
    void openCreate()
  }
)

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  // Expect YYYY-MM-01 or YYYY-MM
  const [y, m] = dateStr.split('-')
  const date = new Date(Number(y), Number(m) - 1, 1)
  return new Intl.DateTimeFormat(locale.value, { month: 'long', year: 'numeric' }).format(date)
}

const statusLabel = (version: BudgetVersion) => {
  return version.is_active ? t('budget.versions.status.active') : t('budget.versions.status.draft')
}

const handleCreate = async () => {
  isSaving.value = true
  try {
    // Build lines payload, filtering out 0%
    const linesToSubmit = initialLines.value
      .filter(l => l.percent > 0)
      .map(l => ({
        category_id: l.categoryId,
        percent: l.percent,
        include_children: false
      }))

    if (linesToSubmit.length === 0) {
      push.error(t('budget.messages.linesRequired'))
      isSaving.value = false
      return
    }

    await createVersion({
      effective_from_month: effectiveMonth.value + '-01',
      lines: linesToSubmit
    })
    push.success(t('budget.messages.versionCreated'))
    formOpen.value = false
  } catch (err) {
    if (isApiError(err)) {
        const codeKey = `common.errors.${err.code}`
        if (te(codeKey)) {
             push.error(t(codeKey))
        } else if (err.code === 'VALIDATION_ERROR') {
             const details = err.details ? JSON.stringify(err.details) : ''
             push.error(`${t('common.errors.VALIDATION_ERROR')}: ${details || err.message}`)
        } else {
             push.error(err.message || t('budget.messages.error'))
        }
    } else {
        push.error(t('budget.messages.error'))
    }
  } finally {
    isSaving.value = false
  }
}

const goToEditor = (version: BudgetVersion) => {
  router.push(`/budgets/editor/${version.id}`)
}

const handleDelete = async (version: BudgetVersion) => {
  if (!confirm(t('budget.confirm.deleteVersion'))) return
  try {
    await deleteVersion(version.id)
    push.success(t('budget.messages.versionDeleted'))
  } catch {
    push.error(t('budget.messages.error'))
  }
}

onMounted(() => {
  fetchVersions()
})
</script>

<template>
  <div class="grid gap-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold">{{ t('budget.title') }}</h1>
        <p class="text-sm text-muted-foreground">{{ t('budget.description') }}</p>
      </div>
      <Button @click="openCreate">
        <Plus class="h-4 w-4 mr-2" />
        {{ t('budget.actions.newVersion') }}
      </Button>
    </div>

    <CrudTableCard
      :title="t('budget.versions.title')"
      :description="t('budget.versions.description')"
      :loading="loading"
      :error="error"
      :empty="versions.length === 0"
      :loading-message="t('budget.loading')"
      :empty-message="t('budget.empty')"
    >
      <div class="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('budget.versions.month') }}</TableHead>
              <TableHead>{{ t('budget.versions.status.label') }}</TableHead>
              <TableHead class="text-right">{{ t('budget.columns.actions') }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="version in versions" :key="version.id">
              <TableCell class="font-medium capitalize">
                {{ formatDate(version.effective_from_month) }}
              </TableCell>
              <TableCell>
                <Badge :variant="version.is_active ? 'default' : 'secondary'">
                  {{ statusLabel(version) }}
                </Badge>
              </TableCell>
              <TableCell class="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="goToEditor(version)">
                      {{ t('budget.actions.edit') }}
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="handleDelete(version)" class="text-destructive">
                      {{ t('budget.actions.delete') }}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </CrudTableCard>

    <Sheet v-model:open="formOpen">
      <SheetContent side="right" class="w-full sm:max-w-md px-6 py-6">
        <SheetHeader>
          <SheetTitle>{{ t('budget.form.createTitle') }}</SheetTitle>
          <SheetDescription>{{ t('budget.form.createDescription') }}</SheetDescription>
        </SheetHeader>
        <div class="mt-6 grid gap-6">
          <div class="grid gap-2">
            <Label>{{ t('budget.form.effectiveMonth') }}</Label>
            <Input v-model="effectiveMonth" type="month" />
          </div>

          <div class="grid gap-2">
             <Label>{{ t('budget.form.initialAllocation') }}</Label>
             <p class="text-xs text-muted-foreground mb-2">{{ t('budget.form.initialAllocationHint') }}</p>
             
             <div v-if="initialLines.length === 0" class="text-sm text-yellow-600 space-y-2">
                <p>{{ t('budget.form.noCategories') }}</p>
                <div class="p-2 bg-slate-100 dark:bg-slate-900 rounded text-xs font-mono">
                  DEBUG: Cats={{ categories.length }}, Out/Req={{ categories.filter(c => c.direction==='out' && c.is_budget_relevant).length }}
                  <br>
                  Sample: {{ categories.slice(0,3).map(c => `${c.name}(${c.direction})`).join(', ') }}
                </div>
             </div>

             <div v-else class="space-y-3 border rounded-md p-3 max-h-[300px] overflow-y-auto">
                <div v-for="(line, idx) in initialLines" :key="line.categoryId" class="flex items-center justify-between">
                   <span class="text-sm font-medium">{{ line.name }}</span>
                   <div class="flex items-center gap-1">
                      <Input 
                        type="number" 
                        v-model.number="line.percent" 
                        class="h-8 w-16 text-right"
                        min="0"
                        max="100"
                      />
                      <span class="text-sm text-muted-foreground">%</span>
                   </div>
                </div>
             </div>
          </div>
        </div>
        <SheetFooter class="mt-6 flex-row justify-end gap-2">
          <Button :disabled="isSaving" @click="handleCreate">
            {{ isSaving ? t('budget.form.saving') : t('budget.form.create') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  </div>
</template>
