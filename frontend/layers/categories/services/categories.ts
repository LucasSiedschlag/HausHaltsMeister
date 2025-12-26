import type { Category, CreateCategoryRequest, UpdateCategoryRequest } from '../types/category'
import { useApiClient } from '~/layers/shared/utils/api'

export function useCategoriesService() {
  const { request } = useApiClient()

  const listCategories = async (activeOnly = true, month?: string) => {
    return request<Category[]>('/categories', {
      query: { active: activeOnly, month },
    })
  }

  const createCategory = async (payload: CreateCategoryRequest) => {
    return request<Category>('/categories', {
      method: 'POST',
      body: payload,
    })
  }

  const deactivateCategory = async (id: number, effectiveMonth?: string) => {
    return request<{ status: string }>(`/categories/${id}/deactivate`, {
      method: 'PATCH',
      query: effectiveMonth ? { effective_month: effectiveMonth } : undefined,
    })
  }

  const updateCategory = async (id: number, payload: UpdateCategoryRequest) => {
    return request<Category>(`/categories/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  return {
    listCategories,
    createCategory,
    updateCategory,
    deactivateCategory,
  }
}
