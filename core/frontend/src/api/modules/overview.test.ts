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

import {
  getOverviewInfo,
  getFailedList,
  getSendQueueList,
  getSendQueueInfo,
  resendQueue,
  deleteSendQueue,
  getSendQueueConfig,
  setSendQueueConfig,
  clearSendQueue,
} from './overview'

describe('overview API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getOverviewInfo calls GET /overview with params', () => {
    const params = { domain: 'test.com', start_time: 1000, end_time: 2000 }
    getOverviewInfo(params)
    expect(mockGet).toHaveBeenCalledWith('/overview', { params })
  })

  it('getFailedList calls GET /overview/failed', () => {
    const params = { domain: 'test.com', start_time: 1000, end_time: 2000 }
    getFailedList(params)
    expect(mockGet).toHaveBeenCalledWith('/overview/failed', { params })
  })

  it('getSendQueueList calls GET /postfix_queue/list', () => {
    getSendQueueList()
    expect(mockGet).toHaveBeenCalledWith('/postfix_queue/list')
  })

  it('getSendQueueInfo calls GET /postfix_queue/queue_info', () => {
    getSendQueueInfo({ queue_id: 'abc123' })
    expect(mockGet).toHaveBeenCalledWith('/postfix_queue/queue_info', { params: { queue_id: 'abc123' } })
  })

  it('resendQueue calls POST /postfix_queue/flush_by_id', () => {
    resendQueue({ queue_ids: ['a', 'b'] })
    expect(mockPost).toHaveBeenCalledWith('/postfix_queue/flush_by_id', { queue_ids: ['a', 'b'] }, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('deleteSendQueue calls POST /postfix_queue/delete_by_id', () => {
    deleteSendQueue({ queue_ids: ['a'] })
    expect(mockPost).toHaveBeenCalledWith('/postfix_queue/delete_by_id', { queue_ids: ['a'] }, expect.any(Object))
  })

  it('getSendQueueConfig calls GET /postfix_queue/get_config', () => {
    getSendQueueConfig()
    expect(mockGet).toHaveBeenCalledWith('/postfix_queue/get_config')
  })

  it('setSendQueueConfig calls POST /postfix_queue/set_all_config', () => {
    const params = { key1: 'val1' }
    setSendQueueConfig(params)
    expect(mockPost).toHaveBeenCalledWith('/postfix_queue/set_all_config', params, expect.any(Object))
  })

  it('clearSendQueue calls POST /postfix_queue/delete with empty default', () => {
    clearSendQueue()
    expect(mockPost).toHaveBeenCalledWith('/postfix_queue/delete', {}, expect.any(Object))
  })

  it('clearSendQueue accepts custom params', () => {
    clearSendQueue({ custom: true })
    expect(mockPost).toHaveBeenCalledWith('/postfix_queue/delete', { custom: true }, expect.any(Object))
  })
})
