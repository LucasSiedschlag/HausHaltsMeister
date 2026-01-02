type InputStyleOptions = {
  touched: boolean
  hasError: boolean
}

export const getInputClass = ({ touched, hasError }: InputStyleOptions) => {
  if (touched && hasError) {
    return [
      'border-destructive/40',
      'focus-visible:outline',
      'focus-visible:outline-2',
      'focus-visible:outline-destructive/60',
      'focus-visible:outline-offset-0',
      'focus-visible:ring-2',
      'focus-visible:ring-destructive/20'
    ].join(' ')
  }
  return [
    'border-primary/30',
    'focus-visible:outline',
    'focus-visible:outline-2',
    'focus-visible:outline-primary/60',
    'focus-visible:outline-offset-0',
    'focus-visible:ring-2',
    'focus-visible:ring-primary/20'
  ].join(' ')
}
