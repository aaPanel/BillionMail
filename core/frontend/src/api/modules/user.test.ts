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

import { login, logout, getValidateCode } from './user'

describe('user API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('login calls POST /login with credentials', () => {
    const params = { username: 'admin', password: 'pass' }
    login(params)
    expect(mockPost).toHaveBeenCalledWith('/login', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('logout calls POST /logout with empty body', () => {
    logout()
    expect(mockPost).toHaveBeenCalledWith('/logout', {}, expect.objectContaining({
      fetchOptions: expect.objectContaining({
        loading: 'user.api.loading.logout',
        successMessage: true,
      }),
    }))
  })

  it('getValidateCode calls GET /get_validate_code', () => {
    getValidateCode()
    expect(mockGet).toHaveBeenCalledWith('/get_validate_code')
  })
})
