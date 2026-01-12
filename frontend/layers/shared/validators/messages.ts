export const messages = {
  required: (label: string) => `${label} é obrigatório.`,
  minLength: (label: string, min: number) => `${label} deve ter pelo menos ${min} caracteres.`,
  maxLength: (label: string, max: number) => `${label} deve ter no máximo ${max} caracteres.`,
  amountPositive: (label: string) => `${label} deve ser maior que zero.`,
  invalidNumber: (label: string) => `${label} deve ser um número válido.`,
  minValue: (label: string, min: number) => `${label} deve ser maior ou igual a ${min}.`,
  maxValue: (label: string, max: number) => `${label} deve ser menor ou igual a ${max}.`,
  invalidCode: (label: string) => `${label} deve conter apenas letras, números e hífen.`,
  invalidOption: (label: string) => `${label} inválido.`,
  invalidDate: (label: string) => `${label} inválida.`,
  emailInvalid: () => 'E-mail inválido.',
  emailNoSpaces: () => 'O e-mail não pode conter espaços.',
  passwordUppercase: () => 'A senha deve conter ao menos uma letra maiúscula.',
  passwordNumber: () => 'A senha deve conter ao menos um número.'
}
