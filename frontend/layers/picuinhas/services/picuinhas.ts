import type {
  AddEntryRequest,
  CreatePersonRequest,
  CreatePicuinhaCaseRequest,
  PicuinhaCase,
  PicuinhaCaseInstallment,
  PaymentMethod,
  PicuinhaEntry,
  Person,
  UpdatePersonRequest,
  UpdatePicuinhaInstallmentRequest,
} from '../types/picuinha'
import type { Category } from '~/layers/categories/types/category'
import { useApiClient } from '~/layers/shared/utils/api'

export function usePicuinhasService() {
  const { request } = useApiClient()

  const listPersons = async () => {
    return request<Person[]>('/picuinhas/persons')
  }

  const createPerson = async (payload: CreatePersonRequest) => {
    return request<Person>('/picuinhas/persons', {
      method: 'POST',
      body: payload,
    })
  }

  const updatePerson = async (id: number, payload: UpdatePersonRequest) => {
    return request<Person>(`/picuinhas/persons/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  const deletePerson = async (id: number) => {
    return request<{ status: string }>(`/picuinhas/persons/${id}`, {
      method: 'DELETE',
    })
  }

  const addEntry = async (payload: AddEntryRequest) => {
    return request<PicuinhaEntry>('/picuinhas/entries', {
      method: 'POST',
      body: payload,
    })
  }

  const listPaymentMethods = async () => {
    return request<PaymentMethod[]>('/payment-methods')
  }

  const listEntries = async () => {
    return request<PicuinhaEntry[]>('/picuinhas/entries')
  }

  const listCases = async (personId: number) => {
    return request<PicuinhaCase[]>(`/picuinhas/cases?person_id=${personId}`)
  }

  const createCase = async (payload: CreatePicuinhaCaseRequest) => {
    return request<PicuinhaCase>('/picuinhas/cases', {
      method: 'POST',
      body: payload,
    })
  }

  const updateCase = async (id: number, payload: CreatePicuinhaCaseRequest) => {
    return request<PicuinhaCase>(`/picuinhas/cases/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  const deleteCase = async (id: number) => {
    return request<{ status: string }>(`/picuinhas/cases/${id}`, {
      method: 'DELETE',
    })
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

  const listCategories = async () => {
    return request<Category[]>('/categories?active=true')
  }

  const updateEntry = async (id: number, payload: AddEntryRequest) => {
    return request<PicuinhaEntry>(`/picuinhas/entries/${id}`, {
      method: 'PUT',
      body: payload,
    })
  }

  const deleteEntry = async (id: number) => {
    return request<{ status: string }>(`/picuinhas/entries/${id}`, {
      method: 'DELETE',
    })
  }

  return {
    listPersons,
    createPerson,
    updatePerson,
    deletePerson,
    addEntry,
    listPaymentMethods,
    listEntries,
    listCases,
    createCase,
    updateCase,
    deleteCase,
    listCaseInstallments,
    updateCaseInstallment,
    listCategories,
    updateEntry,
    deleteEntry,
  }
}
