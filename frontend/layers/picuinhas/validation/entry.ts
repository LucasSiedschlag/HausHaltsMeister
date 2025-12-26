import type { AddEntryRequest, CardOwner, PicuinhaKind } from '../types/picuinha'

export interface EntryFormErrors {
  person_id?: string
  amount?: string
  kind?: string
}

const allowedKinds: PicuinhaKind[] = ['PLUS', 'MINUS']
const allowedOwners: CardOwner[] = ['SELF', 'THIRD']

export function validateEntryInput(input: AddEntryRequest) {
  const errors: EntryFormErrors = {}

  if (!input.person_id) {
    errors.person_id = 'Selecione a pessoa.'
  }

  if (input.amount <= 0) {
    errors.amount = 'Informe um valor maior que zero.'
  }

  if (!allowedKinds.includes(input.kind)) {
    errors.kind = 'Selecione o tipo correto.'
  }

  if (input.card_owner && !allowedOwners.includes(input.card_owner)) {
    errors.kind = 'Tipo de cartão inválido.'
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
  }
}
