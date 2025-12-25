import type { Category, CreateCategoryRequest } from '../types/category'
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

  // TODO: Backend não expõe endpoint para atualizar categorias. Necessário adicionar PATCH/PUT.
  return {
    listCategories,
    createCategory,
    deactivateCategory,
  }
}
