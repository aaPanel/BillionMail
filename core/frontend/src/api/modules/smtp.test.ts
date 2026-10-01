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

import { getSmtpList, addSmtp, editSmtp, testSmtp, deleteSmtp, getUnbindDomains } from './smtp'

describe('smtp API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getSmtpList calls GET /relay/list', () => {
    getSmtpList()
    expect(mockGet).toHaveBeenCalledWith('/relay/list')
  })

  it('addSmtp calls POST /relay/add with fetchOptions', () => {
    const params = {
      rtype: 'smtp',
      remark: 'test',
      smtp_name: 'Test SMTP',
      sender_domains: ['test.com'],
      relay_host: 'smtp.test.com',
      relay_port: 587,
      auth_user: 'user',
      auth_password: 'pass',
      active: 1,
    }
    addSmtp(params)
    expect(mockPost).toHaveBeenCalledWith('/relay/add', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({
        successMessage: true,
        loading: 'smtp.api.loading.adding',
      }),
    }))
  })

  it('editSmtp calls POST /relay/edit', () => {
    const params = { id: 1, smtp_name: 'Updated' }
    editSmtp(params)
    expect(mockPost).toHaveBeenCalledWith('/relay/edit', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('testSmtp calls POST /relay/test_connection', () => {
    const params = {
      sender_domains: ['test.com'],
      relay_host: 'smtp.test.com',
      relay_port: 587,
      auth_user: 'user',
      auth_password: 'pass',
    }
    testSmtp(params)
    expect(mockPost).toHaveBeenCalledWith('/relay/test_connection', params, expect.any(Object))
  })

  it('deleteSmtp calls POST /relay/delete', () => {
    deleteSmtp({ id: 5 })
    expect(mockPost).toHaveBeenCalledWith('/relay/delete', { id: 5 }, expect.any(Object))
  })

  it('getUnbindDomains calls GET /relay/get_unbound_domains', () => {
    getUnbindDomains()
    expect(mockGet).toHaveBeenCalledWith('/relay/get_unbound_domains')
  })
})
