import type { PaymentMethod } from '../types/payment-method'
import type { CreateInstallmentRequest, InstallmentCategoryOption, InstallmentPlanResponse, InvoiceSummary } from '../types/installment'
import { useApiClient } from '~/layers/shared/utils/api'

export function useInstallmentsService() {
  const { request } = useApiClient()

  const listPaymentMethods = async () => {
    return request<PaymentMethod[]>('/payment-methods')
  }

  const listCategories = async () => {
    return request<InstallmentCategoryOption[]>('/categories', {
      query: { active: true },
    })
  }

  const createInstallment = async (payload: CreateInstallmentRequest) => {
    return request<InstallmentPlanResponse>('/installments', {
      method: 'POST',
      body: payload,
    })
  }

  const getInvoice = async (paymentMethodId: number, month: string) => {
    return request<InvoiceSummary>(`/payment-methods/${paymentMethodId}/invoice`, {
      query: { month },
    })
  }

  return {
    listPaymentMethods,
    listCategories,
    createInstallment,
    getInvoice,
  }
}
