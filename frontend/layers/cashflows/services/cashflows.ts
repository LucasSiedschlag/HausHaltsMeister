import type { CashflowEntry, CreateCashflowRequest, Direction } from '../types/cashflow'
import type { Category } from '~/layers/categories/types/category'
import type { PaymentMethod } from '~/layers/payment-methods/types/payment-method'
import { useApiClient } from '~/layers/shared/utils/api'

export function useCashflowsService() {
  const { request } = useApiClient()

  const listCashflows = async (month: string, filters: { direction?: Direction; is_fixed?: boolean } = {}) => {
    return request<CashflowEntry[]>('/cashflows', {
      query: {
        month,
        direction: filters.direction,
        is_fixed: typeof filters.is_fixed === 'boolean' ? filters.is_fixed : undefined,
      },
    })
  }

  const createCashflow = async (payload: CreateCashflowRequest) => {
    return request<CashflowEntry>('/cashflows', {
      method: 'POST',
      body: payload,
    })
  }

  const updateCashflow = async (id: number, payload: CreateCashflowRequest) => {
    return request<CashflowEntry>(`/cashflows/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  const deleteCashflow = async (id: number) => {
    return request<{ status: string }>(`/cashflows/${id}`, {
      method: 'DELETE',
    })
  }

  const reverseCashflow = async (id: number) => {
    return request<CashflowEntry>(`/cashflows/${id}/reverse`, {
      method: 'POST',
    })
  }

  const copyFixedCashflows = async (fromMonth: string, toMonth: string) => {
    return request<{ copied_count: number }>('/cashflows/copy-fixed', {
      method: 'POST',
      body: { from_month: fromMonth, to_month: toMonth },
    })
  }

  const listCategories = async (activeOnly = true, month?: string) => {
    return request<Category[]>('/categories', {
      query: { active: activeOnly, month },
    })
  }

  const listPaymentMethods = async () => {
    return request<PaymentMethod[]>('/payment-methods')
  }

  return {
    listCashflows,
    createCashflow,
    updateCashflow,
    deleteCashflow,
    reverseCashflow,
    copyFixedCashflows,
    listCategories,
    listPaymentMethods,
  }
}
