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

import { getForwardList, addForward, editForward, deleteForward } from './forward'

describe('settings/forward API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getForwardList calls GET /mail_forward/list', () => {
    const params = { page: 1, page_size: 10, domain: 'test.com' }
    getForwardList(params)
    expect(mockGet).toHaveBeenCalledWith('/mail_forward/list', { params })
  })

  it('addForward calls POST /mail_forward/add', () => {
    const params = { active: 1, address: 'a@test.com', domain: 'test.com', goto: 'b@other.com' }
    addForward(params)
    expect(mockPost).toHaveBeenCalledWith('/mail_forward/add', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('editForward calls POST /mail_forward/edit', () => {
    const params = { active: 1, address: 'a@test.com', goto: 'b@other.com' }
    editForward(params)
    expect(mockPost).toHaveBeenCalledWith('/mail_forward/edit', params, expect.any(Object))
  })

  it('deleteForward calls POST /mail_forward/delete', () => {
    deleteForward({ address: 'a@test.com' })
    expect(mockPost).toHaveBeenCalledWith('/mail_forward/delete', { address: 'a@test.com' }, expect.any(Object))
  })
})
