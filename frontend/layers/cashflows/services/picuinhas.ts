import type { Person, PicuinhaCase, PicuinhaCaseInstallment, UpdatePicuinhaInstallmentRequest } from '~/layers/picuinhas/types/picuinha'
import { useApiClient } from '~/layers/shared/utils/api'

export function usePicuinhasMonthlyService() {
  const { request } = useApiClient()

  const listPersons = async () => {
    return request<Person[]>('/picuinhas/persons')
  }

  const listCases = async (personId: number) => {
    return request<PicuinhaCase[]>(`/picuinhas/cases?person_id=${personId}`)
  }

  const listCaseInstallments = async (caseId: number) => {
    return request<PicuinhaCaseInstallment[]>(`/picuinhas/cases/${caseId}/installments`)
  }

  const updateCaseInstallment = async (id: number, payload: UpdatePicuinhaInstallmentRequest) => {
    return request<PicuinhaCaseInstallment>(`/picuinhas/installments/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  return {
    listPersons,
    listCases,
    listCaseInstallments,
    updateCaseInstallment,
  }
}
