import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()

vi.mock('@/api', () => ({
  instance: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
  },
}))

import { getScraperStatus, startScrapeJob } from './enrichment'

describe('enrichment API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getScraperStatus calls GET /frostbyte/proxy/leads/pipeline', () => {
    getScraperStatus()
    expect(mockGet).toHaveBeenCalledWith('/frostbyte/proxy/leads/pipeline')
  })

  it('startScrapeJob calls POST /frostbyte/proxy/leads/scrape', () => {
    const data = { query: 'restaurants', location: 'NYC', max_results: 50 }
    startScrapeJob(data)
    expect(mockPost).toHaveBeenCalledWith('/frostbyte/proxy/leads/scrape', data)
  })
})
