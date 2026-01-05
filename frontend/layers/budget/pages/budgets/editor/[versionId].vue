<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@shared/components/ui/button'
import { Badge } from '@shared/components/ui/badge'
import { Input } from '@shared/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@shared/components/ui/table'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@shared/components/ui/card'
import { CornerDownRight, ArrowLeft } from 'lucide-vue-next'
import { useCategories, type Category } from '#layers/categories/composables/useCategories'
import { useBudgetEditor, type BudgetLine } from '../../../composables/useBudgetEditor'
import { useBudget, type BudgetVersion } from '../../../composables/useBudget'
import { useI18n } from 'vue-i18n'
import { push } from 'notivue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const versionId = route.params.versionId as string

const { categories, fetchCategories } = useCategories()
const { lines, loading: linesLoading, fetchLines, updateLine, createLine } = useBudgetEditor()
const { versions, fetchVersions } = useBudget()

const currentVersion = computed(() => versions.value.find((v: BudgetVersion) => v.id === versionId))

// State for manual calculation base since backend doesn't store it yet
const totalEstimatedIncome = ref(0)
const isSaving = ref(false)

type BudgetRow = Category & {
  depth: number
  lineId?: string
  percent: number
  plannedAmount: number
  actualAmount: number // Mocked/Future
  balance: number // Mocked/Future
}

// Hierarchy helper
const buildHierarchy = (items: Category[], budgetLines: BudgetLine[], incomeBase: number) => {
  const lineMap = new Map(budgetLines.map(l => [l.category_id, l]))
  const sorted = [...items]
    .filter(c => c.is_budget_relevant && c.direction === 'out') // Only OUT categories are budgeted? Legacy implies this.
    .sort((a, b) => a.name.localeCompare(b.name))

  const byParent = new Map<string | null, Category[]>()
  
  for (const item of sorted) {
    const key = item.parent_id ?? null
    const list = byParent.get(key) ?? []
    list.push(item)
    byParent.set(key, list)
  }

  const walk = (parentKey: string | null, depth: number, acc: BudgetRow[]) => {
    const children = byParent.get(parentKey) ?? []
    for (const child of children) {
      const line = lineMap.get(child.id)
      const percent = line?.percent || 0
      const planned = incomeBase * (percent / 100)

      acc.push({
        ...child,
        depth,
        lineId: line?.id,
        percent,
        plannedAmount: planned,
        actualAmount: 0, // Not available yet
        balance: planned - 0
      })
      walk(child.id, depth + 1, acc)
    }
  }

  const output: BudgetRow[] = []
  walk(null, 0, output)
  return output
}

const budgetRows = computed<BudgetRow[]>(() => {
  if (!categories.value.length) return []
  return buildHierarchy(categories.value, lines.value, totalEstimatedIncome.value)
})

const totalPercent = computed(() => {
  return lines.value.reduce((acc: number, l: BudgetLine) => acc + l.percent, 0)
})

const totalPlanned = computed(() => {
  return totalEstimatedIncome.value * (totalPercent.value / 100)
})

const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

function formatCurrency(value: number) {
  return currency.format(value || 0)
}

function deltaBadgeClass(planned: number, actual: number) {
  const delta = planned - actual
  if (delta >= 0) {
    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/40 dark:bg-emerald-500/20 dark:text-emerald-300'
  }
  return 'border-red-500/30 bg-red-500/10 text-red-700 dark:border-red-500/40 dark:bg-red-500/20 dark:text-red-300'
}

const handlePercentChange = async (row: BudgetRow, value: string) => {
  let newPercent = Number(value)
  if (isNaN(newPercent)) newPercent = 0
  if (newPercent < 0) newPercent = 0
  if (newPercent > 100) newPercent = 100

  // Optimistic update would be nice, but we rely on store refresh for now
  try {
    if (row.lineId) {
       await updateLine(versionId, row.lineId, { percent: newPercent })
    } else if (newPercent > 0) {
       await createLine(versionId, row.id, newPercent)
    }
    // Refresh happens automatically if composable is reactive? 
    await fetchLines(versionId) 
  } catch {
    push.error(t('budget.messages.error'))
  }
}

onMounted(async () => {
  await Promise.all([
    fetchCategories({ active: true, direction: 'out' }),
    fetchLines(versionId),
    fetchVersions() 
  ])
  // Default estimated income for demo purposes if 0
  if (totalEstimatedIncome.value === 0) totalEstimatedIncome.value = 5000
})
</script>

<template>
  <div class="grid gap-6">
    <div class="flex items-center gap-4">
      <Button variant="ghost" size="icon" @click="router.back()">
        <ArrowLeft class="h-4 w-4" />
      </Button>
      <div>
        <h1 class="text-2xl font-semibold">{{ t('budget.editor.title') }}</h1>
        <p class="text-sm text-muted-foreground" v-if="currentVersion">
           {{ t('budget.editor.subtitle', { month: currentVersion.effective_from_month }) }}
        </p>
      </div>
      <div class="ml-auto flex items-center gap-4">
          <!-- Total Income Input Helper -->
         <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-muted-foreground">Renda Base:</span>
            <Input 
                type="number" 
                v-model="totalEstimatedIncome" 
                class="w-28 text-right font-mono"
            />
         </div>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid gap-4 md:grid-cols-3">
        <Card>
            <CardHeader class="pb-2">
                <CardTitle class="text-sm font-medium text-muted-foreground">Volume Planejado</CardTitle>
            </CardHeader>
            <CardContent>
                <div class="text-2xl font-bold">{{ formatCurrency(totalPlanned) }}</div>
                <p class="text-xs text-muted-foreground">
                    Baseado em {{ totalPercent }}% da renda
                </p>
            </CardContent>
        </Card>
        <Card>
            <CardHeader class="pb-2">
                <CardTitle class="text-sm font-medium text-muted-foreground">Distribuição</CardTitle>
            </CardHeader>
            <CardContent>
                <div class="text-2xl font-bold" :class="totalPercent > 100 ? 'text-destructive' : 'text-primary'">
                    {{ totalPercent }}%
                </div>
                <p class="text-xs text-muted-foreground">
                    {{ totalPercent > 100 ? 'Orçamento estourado!' : 'Dentro do limite' }}
                </p>
            </CardContent>
        </Card>
        <Card>
             <CardHeader class="pb-2">
                <CardTitle class="text-sm font-medium text-muted-foreground">Status</CardTitle>
            </CardHeader>
            <CardContent>
                <div class="text-2xl font-bold text-muted-foreground">
                    {{ isSaving ? 'Salvando...' : 'Salvo' }}
                </div>
                <p class="text-xs text-muted-foreground">
                    Alterações salvas automaticamente
                </p>
            </CardContent>
        </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>{{ t('budget.editor.categoriesTitle') }}</CardTitle>
        <CardDescription>{{ t('budget.editor.categoriesDescription') }}</CardDescription>
      </CardHeader>
      <CardContent class="p-0">
          <div class="overflow-x-auto">
           <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-[300px]">{{ t('budget.columns.category') }}</TableHead>
                <TableHead class="w-[150px]">{{ t('budget.columns.percent') }}</TableHead>
                <TableHead>{{ t('budget.columns.planned') }}</TableHead>
                <TableHead>{{ t('budget.columns.actual') }}</TableHead>
                <TableHead>{{ t('budget.columns.balance') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="row in budgetRows" :key="row.id">
                <TableCell>
                  <div class="flex items-center gap-2" :style="{ paddingLeft: `${row.depth * 16}px` }">
                    <CornerDownRight v-if="row.depth > 0" class="h-3.5 w-3.5 text-muted-foreground" />
                    <span class="font-medium">{{ row.name }}</span>
                    <Badge v-if="!row.is_active" variant="outline" class="text-xs">Inativa</Badge>
                  </div>
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-1">
                    <Input
                      type="number"
                      class="h-8 w-16 text-right"
                      min="0"
                      max="100"
                      step="1"
                      :model-value="row.percent"
                      @change="(e: Event) => handlePercentChange(row, (e.target as HTMLInputElement).value)"
                    />
                    <span class="text-sm text-muted-foreground">%</span>
                  </div>
                </TableCell>
                <TableCell>{{ formatCurrency(row.plannedAmount) }}</TableCell>
                <TableCell class="text-muted-foreground">{{ formatCurrency(row.actualAmount) }}</TableCell>
                <TableCell>
                    <Badge variant="outline" :class="deltaBadgeClass(row.plannedAmount, row.actualAmount)">
                        {{ formatCurrency(row.balance) }}
                    </Badge>
                </TableCell>
              </TableRow>
              <TableRow v-if="budgetRows.length === 0">
                 <TableCell colspan="5" class="h-24 text-center">
                    {{ linesLoading ? t('budget.loading') : t('budget.emptyCategories') }}
                 </TableCell>
              </TableRow>
            </TableBody>
          </Table>
         </div>
      </CardContent>
    </Card>
  </div>
</template>
