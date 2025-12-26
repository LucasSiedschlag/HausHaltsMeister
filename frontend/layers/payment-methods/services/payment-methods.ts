import type { CreatePaymentMethodRequest, PaymentMethod, UpdatePaymentMethodRequest } from '../types/payment-method'
import { useApiClient } from '~/layers/shared/utils/api'

export function usePaymentMethodsService() {
  const { request } = useApiClient()

  const listPaymentMethods = async () => {
    return request<PaymentMethod[]>('/payment-methods')
  }

  const createPaymentMethod = async (payload: CreatePaymentMethodRequest) => {
    return request<PaymentMethod>('/payment-methods', {
      method: 'POST',
      body: payload,
    })
  }

  const updatePaymentMethod = async (id: number, payload: UpdatePaymentMethodRequest) => {
    return request<PaymentMethod>(`/payment-methods/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  const deletePaymentMethod = async (id: number) => {
    return request<{ status: string }>(`/payment-methods/${id}`, {
      method: 'DELETE',
    })
  }

  return {
    listPaymentMethods,
    createPaymentMethod,
    updatePaymentMethod,
    deletePaymentMethod,
  }
}
