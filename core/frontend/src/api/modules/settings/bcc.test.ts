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

import { getBccList, addBcc, editBcc, deleteBcc } from './bcc'

describe('settings/bcc API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getBccList calls GET /mail_bcc/list', () => {
    const params = { page: 1, page_size: 10, domain: 'test.com' }
    getBccList(params)
    expect(mockGet).toHaveBeenCalledWith('/mail_bcc/list', { params })
  })

  it('getBccList works with optional params', () => {
    const params = { page: 1, page_size: 10 }
    getBccList(params)
    expect(mockGet).toHaveBeenCalledWith('/mail_bcc/list', { params })
  })

  it('addBcc calls POST /mail_bcc/add', () => {
    const params = { type: 'sender', address: 'a@test.com', goto: 'b@test.com', active: 1 }
    addBcc(params)
    expect(mockPost).toHaveBeenCalledWith('/mail_bcc/add', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('editBcc calls POST /mail_bcc/edit', () => {
    const params = { id: 1, type: 'sender', address: 'a@test.com', goto: 'b@test.com', active: 1 }
    editBcc(params)
    expect(mockPost).toHaveBeenCalledWith('/mail_bcc/edit', params, expect.any(Object))
  })

  it('deleteBcc calls POST /mail_bcc/delete', () => {
    deleteBcc({ id: 5 })
    expect(mockPost).toHaveBeenCalledWith('/mail_bcc/delete', { id: 5 }, expect.any(Object))
  })
})
