import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import type { FetchError } from 'ofetch'
import { normalizeApiError } from '../layers/shared/utils/api-error'

describe('normalizeApiError', () => {
  it('preserves code, status, and details for rate limits', () => {
    const fetchError = {
      name: 'FetchError',
      message: 'Too many requests',
      status: 429,
      data: {
        code: 'RATE_LIMITED',
        message: 'Slow down',
        details: { retry_after: '10s' }
      }
    } as unknown as FetchError

    const normalized = normalizeApiError(fetchError)
    assert.equal(normalized.code, 'RATE_LIMITED')
    assert.equal(normalized.status, 429)
    assert.deepEqual(normalized.details, { retry_after: '10s' })
  })
})
