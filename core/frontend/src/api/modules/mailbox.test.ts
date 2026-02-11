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
  getMailboxList,
  createMailbox,
  createBatchMailbox,
  updateMailbox,
  deleteMailbox,
  exportMailbox,
  importMailbox,
} from './mailbox'

describe('mailbox API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getMailboxList calls GET /mailbox/list with params', () => {
    const params = { page: 1, page_size: 20, domain: 'test.com', keyword: 'user' }
    getMailboxList(params)
    expect(mockGet).toHaveBeenCalledWith('/mailbox/list', { params })
  })

  it('getMailboxList works with null domain', () => {
    const params = { page: 1, page_size: 20, domain: null }
    getMailboxList(params)
    expect(mockGet).toHaveBeenCalledWith('/mailbox/list', { params })
  })

  it('createMailbox calls POST /mailbox/create', () => {
    const params = {
      full_name: 'user@test.com',
      domain: 'test.com',
      password: 'pass123',
      active: 1,
      isAdmin: 0,
      quota: 1024,
    }
    createMailbox(params)
    expect(mockPost).toHaveBeenCalledWith('/mailbox/create', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({
        loading: 'mailbox.api.loading.creating',
        successMessage: true,
      }),
    }))
  })

  it('createBatchMailbox calls POST /mailbox/batch_create', () => {
    const params = { domain: 'test.com', prefix: 'user', count: 5, quota: 1024 }
    createBatchMailbox(params)
    expect(mockPost).toHaveBeenCalledWith('/mailbox/batch_create', params, expect.any(Object))
  })

  it('updateMailbox calls POST /mailbox/update', () => {
    const params = {
      full_name: 'user@test.com',
      domain: 'test.com',
      password: 'newpass',
      active: 1,
      isAdmin: 0,
      quota: 2048,
    }
    updateMailbox(params)
    expect(mockPost).toHaveBeenCalledWith('/mailbox/update', params, expect.any(Object))
  })

  it('deleteMailbox calls POST /mailbox/delete', () => {
    deleteMailbox({ emails: ['a@test.com', 'b@test.com'] })
    expect(mockPost).toHaveBeenCalledWith('/mailbox/delete', { emails: ['a@test.com', 'b@test.com'] }, expect.any(Object))
  })

  it('exportMailbox calls POST /mailbox/export with blob responseType', () => {
    const params = { domain: 'test.com', file_type: 'csv' }
    exportMailbox(params)
    expect(mockPost).toHaveBeenCalledWith('/mailbox/export', params, expect.objectContaining({
      responseType: 'blob',
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('importMailbox calls POST /mailbox/import', () => {
    const params = { file_data: 'base64data', file_type: 'csv' }
    importMailbox(params)
    expect(mockPost).toHaveBeenCalledWith('/mailbox/import', params, expect.any(Object))
  })
})
