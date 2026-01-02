export const messages = {
  required: (label: string) => `${label} é obrigatório.`,
  minLength: (label: string, min: number) => `${label} deve ter pelo menos ${min} caracteres.`,
  maxLength: (label: string, max: number) => `${label} deve ter no máximo ${max} caracteres.`,
  emailInvalid: () => 'E-mail inválido.',
  emailNoSpaces: () => 'O e-mail não pode conter espaços.',
  passwordUppercase: () => 'A senha deve conter ao menos uma letra maiúscula.',
  passwordNumber: () => 'A senha deve conter ao menos um número.'
}
