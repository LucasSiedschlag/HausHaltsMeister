import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { hasRoleRank, roleRank } from '../layers/shared/utils/ledger-roles'

describe('ledger role helpers', () => {
  it('orders roles from viewer -> editor -> owner', () => {
    assert(roleRank('owner') > roleRank('editor'))
    assert(roleRank('editor') > roleRank('viewer'))
  })

  it('hasRoleRank respects the required threshold', () => {
    assert(hasRoleRank('owner', 'editor'))
    assert(hasRoleRank('editor', 'viewer'))
    assert(!hasRoleRank('viewer', 'editor'))
    assert(!hasRoleRank(null, 'viewer'))
  })
})
