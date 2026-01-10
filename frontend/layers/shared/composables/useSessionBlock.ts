export const useSessionBlock = () => {
  const blocked = useState<boolean>('session_expired_blocked', () => false)

  return { blocked }
}
