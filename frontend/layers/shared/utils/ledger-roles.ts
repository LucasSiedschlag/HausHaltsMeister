export type LedgerRole = 'owner' | 'editor' | 'viewer'

const ROLE_RANKS: Record<LedgerRole, number> = {
  viewer: 1,
  editor: 2,
  owner: 3
}

export const roleRank = (role: LedgerRole | string | null): number => {
  if (!role) return 0
  return ROLE_RANKS[role as LedgerRole] ?? 0
}

export const hasRoleRank = (current: LedgerRole | null, required: LedgerRole): boolean =>
  Boolean(current) && roleRank(current) >= roleRank(required)
