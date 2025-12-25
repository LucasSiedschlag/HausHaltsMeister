import type { Category, CreateCategoryRequest, UpdateCategoryRequest } from '../types/category'
import { useApiClient } from '~/layers/shared/utils/api'

export function useCategoriesService() {
  const { request } = useApiClient()

  const listCategories = async (activeOnly = true) => {
    return request<Category[]>('/categories', {
      query: { active: activeOnly },
    })
  }

  const createCategory = async (payload: CreateCategoryRequest) => {
    return request<Category>('/categories', {
      method: 'POST',
      body: payload,
    })
  }

  const deactivateCategory = async (id: number) => {
    return request<{ status: string }>(`/categories/${id}/deactivate`, {
      method: 'PATCH',
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
