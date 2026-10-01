import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()

vi.mock('@/api', () => ({
  instance: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
  },
}))

vi.mock('@/i18n', () => ({
  default: {
    global: { t: (key: string) => key },
  },
}))

import { getApiList, getOverviewStats, createApi, updateApi, deleteApi, testApi } from './api'

describe('api (Send API) module', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getApiList calls GET /batch_mail/api/list', () => {
    const params = {
      page: 1,
      page_size: 10,
      keyword: '',
      active: 1,
      start_time: 0,
      end_time: 9999,
    }
    getApiList(params)
    expect(mockGet).toHaveBeenCalledWith('/batch_mail/api/list', { params })
  })

  it('getOverviewStats calls GET /batch_mail/api/overview_stats', () => {
    const params = { start_time: 1000, end_time: 2000 }
    getOverviewStats(params)
    expect(mockGet).toHaveBeenCalledWith('/batch_mail/api/overview_stats', { params })
  })

  it('createApi calls POST /batch_mail/api/create', () => {
    const params = {
      api_name: 'Test API',
      template_id: 1,
      subject: 'Hello',
      addresser: 'sender@test.com',
      full_name: 'Sender',
      unsubscribe: 0,
      active: 1,
      ip_whitelist: [],
    }
    createApi(params)
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/api/create', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('updateApi calls POST /batch_mail/api/update', () => {
    const params = {
      id: 1,
      api_name: 'Updated',
      template_id: 2,
      subject: 'Hi',
      addresser: 'sender@test.com',
      full_name: 'Sender',
      unsubscribe: 1,
      active: 1,
      ip_whitelist: ['1.2.3.4'],
    }
    updateApi(params)
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/api/update', params, expect.any(Object))
  })

  it('deleteApi calls POST /batch_mail/api/delete', () => {
    deleteApi({ id: 3 })
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/api/delete', { id: 3 }, expect.any(Object))
  })

  it('testApi calls POST /batch_mail/api/send with x-api-key header', () => {
    testApi('my-api-key', { recipient: 'user@test.com' })
    expect(mockPost).toHaveBeenCalledWith(
      '/batch_mail/api/send',
      { recipient: 'user@test.com' },
      expect.objectContaining({
        headers: { 'x-api-key': 'my-api-key' },
        fetchOptions: expect.objectContaining({ successMessage: true }),
      })
    )
  })
})
