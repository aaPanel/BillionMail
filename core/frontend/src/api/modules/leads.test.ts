import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()

vi.mock('@/api', () => ({
  instance: {
    get: (...args: any[]) => mockGet(...args),
  },
}))

import { getLeadList, getLeadDetail } from './leads'

describe('leads API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getLeadList calls GET /frostbyte/proxy/leads with params', () => {
    const params = { page: 1, page_size: 20, keyword: 'test' }
    getLeadList(params)
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/leads', { params })
  })

  it('getLeadList works without keyword', () => {
    const params = { page: 1, page_size: 10 }
    getLeadList(params)
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/leads', { params })
  })

  it('getLeadDetail calls GET /frostbyte/proxy/leads/:id', () => {
    getLeadDetail(42)
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/leads/42')
  })

  it('getLeadDetail uses id in URL path', () => {
    getLeadDetail(1)
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/leads/1')
  })
})
