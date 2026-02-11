import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()

vi.mock('@/api', () => ({
  instance: {
    get: (...args: any[]) => mockGet(...args),
  },
}))

import { getSequenceList, getSequenceDetail } from './sequences'

describe('sequences API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getSequenceList calls GET /frostbyte/proxy/campaigns with params', () => {
    const params = { page: 1, page_size: 10, keyword: 'campaign' }
    getSequenceList(params)
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/campaigns', { params })
  })

  it('getSequenceList works without keyword', () => {
    const params = { page: 2, page_size: 20 }
    getSequenceList(params)
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/campaigns', { params })
  })

  it('getSequenceDetail calls GET /frostbyte/proxy/campaigns/:id', () => {
    getSequenceDetail('abc-123')
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/campaigns/abc-123')
  })

  it('getSequenceDetail uses string id in URL path', () => {
    getSequenceDetail('xyz')
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/campaigns/xyz')
  })
})
