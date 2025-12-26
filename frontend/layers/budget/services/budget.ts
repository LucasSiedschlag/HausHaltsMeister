import type { BudgetItem, BudgetSummary, CategoryOption, SetBudgetItemRequest, UpdateBudgetItemRequest } from '../types/budget'
import { useApiClient } from '~/layers/shared/utils/api'

export function useBudgetService() {
  const { request } = useApiClient()

  const getSummary = async (month: string) => {
    return request<BudgetSummary>(`/budgets/${month}/summary`)
  }

  const setItem = async (month: string, payload: SetBudgetItemRequest) => {
    return request<BudgetItem>(`/budgets/${month}/items`, {
      method: 'POST',
      body: payload,
    })
  }

  const setItemsBulk = async (month: string, items: Array<{ category_id: number; target_percent: number }>) => {
    return request<{ status: string }>(`/budgets/${month}/items`, {
      method: 'PUT',
      body: { items },
    })
  }

  const updateItem = async (id: number, payload: UpdateBudgetItemRequest) => {
    return request<BudgetItem>(`/budgets/items/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  const listCategories = async (activeOnly = true, month?: string) => {
    return request<CategoryOption[]>('/categories', {
      query: { active: activeOnly, month },
    })
  }

  return {
    getSummary,
    setItem,
    setItemsBulk,
    updateItem,
    listCategories,
  }
}
