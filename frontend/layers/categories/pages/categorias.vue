<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import CategoryTable from '../components/CategoryTable.vue'
import CategoryFormSheet from '../components/CategoryFormSheet.vue'
import CategoryDeleteDialog from '../components/CategoryDeleteDialog.vue'
import type { Category, CreateCategoryRequest, UpdateCategoryRequest } from '../types/category'
import { useCategoriesService } from '../services/categories'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'

definePageMeta({
  layout: 'default',
})

const { listCategories, createCategory, updateCategory, deactivateCategory } = useCategoriesService()

const categories = ref<Category[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const selectedCategory = ref<Category | null>(null)
const formError = ref<string | null>(null)
const submitting = ref(false)

const deleteOpen = ref(false)
const deleteTarget = ref<Category | null>(null)
const deleting = ref(false)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const pageSubtitle = computed(() => 'Cadastro e manutenção das categorias do sistema.')

function monthParamFromDate(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  return `${year}-${month}-01`
}

function getCurrentMonthParam() {
  return monthParamFromDate(new Date())
}

function getNextMonthParam() {
  const now = new Date()
  return monthParamFromDate(new Date(now.getFullYear(), now.getMonth() + 1, 1))
}

async function fetchCategories() {
  loading.value = true
  loadError.value = null
  try {
    categories.value = await listCategories(false)
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  formMode.value = 'create'
  selectedCategory.value = null
  formError.value = null
  formOpen.value = true
}

function openEdit(category: Category) {
  formMode.value = 'edit'
  selectedCategory.value = category
  formError.value = null
  formOpen.value = true
}

async function handleSubmit(payload: UpdateCategoryRequest) {
  submitting.value = true
  formError.value = null
  try {
    if (formMode.value === 'edit' && selectedCategory.value) {
      const updated = await updateCategory(selectedCategory.value.id, payload)
      categories.value = categories.value.map((item) => (item.id === updated.id ? updated : item))
      feedback.value = { type: 'success', message: 'Categoria atualizada com sucesso.' }
    } else {
      const createPayload: CreateCategoryRequest = {
        name: payload.name,
        direction: payload.direction,
        is_budget_relevant: payload.is_budget_relevant,
      }
      const created = await createCategory(createPayload)
      categories.value = [created, ...categories.value]
      feedback.value = { type: 'success', message: 'Categoria criada com sucesso.' }
    }
    formOpen.value = false
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function requestDelete(category: Category) {
  deleteTarget.value = category
  deleteOpen.value = true
}

async function confirmDelete(scope: 'current' | 'next') {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    const effectiveMonth = scope === 'next' ? getNextMonthParam() : getCurrentMonthParam()
    await deactivateCategory(deleteTarget.value.id, effectiveMonth)
    await fetchCategories()
    feedback.value = {
      type: 'success',
      message: scope === 'next' ? 'Categoria desativada a partir do próximo mês.' : 'Categoria desativada para o mês atual.',
    }
    deleteOpen.value = false
  } catch (error) {
    feedback.value = { type: 'error', message: getApiErrorMessage(error) }
  } finally {
    deleting.value = false
  }
}

watch(formOpen, (open) => {
  if (!open) {
    formError.value = null
  }
})

onMounted(fetchCategories)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Categorias</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" @click="fetchCategories">Atualizar lista</Button>
    </div>

    <div
      v-if="feedback"
      class="rounded-md border px-4 py-3 text-sm"
      :class="feedback.type === 'error' ? 'border-destructive/30 bg-destructive/10 text-destructive' : 'border-primary/30 bg-primary/10 text-primary'"
    >
      <div class="flex items-center justify-between gap-4">
        <span>{{ feedback.message }}</span>
        <Button variant="ghost" size="sm" @click="feedback = null">Fechar</Button>
      </div>
    </div>

    <CategoryTable
      :categories="categories"
      :loading="loading"
      :error="loadError"
      @create="openCreate"
      @edit="openEdit"
      @remove="requestDelete"
      @retry="fetchCategories"
    />

    <CategoryFormSheet
      v-model:open="formOpen"
      :mode="formMode"
      :category="selectedCategory"
      :submitting="submitting"
      :error-message="formError"
      @submit="handleSubmit"
    />

    <CategoryDeleteDialog
      v-model:open="deleteOpen"
      :category="deleteTarget"
      :submitting="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>
