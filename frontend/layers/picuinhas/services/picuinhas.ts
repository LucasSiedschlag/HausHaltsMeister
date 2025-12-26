import type {
  AddEntryRequest,
  CreatePersonRequest,
  PaymentMethod,
  PicuinhaEntry,
  Person,
  UpdatePersonRequest,
} from '../types/picuinha'
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
    updateEntry,
    deleteEntry,
  }
}
