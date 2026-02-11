import { describe, it, expect, vi } from 'vitest'

// Mock all external dependencies before importing
vi.mock('@/router', () => ({
  default: { push: vi.fn() },
}))

vi.mock('@/store', () => ({
  useUserStore: vi.fn(() => ({
    login: { token: 'test-token' },
    resetLoginInfo: vi.fn(),
  })),
}))

vi.mock('@/utils', () => ({
  apiUrlPrefix: '/test-prefix',
  isObject: (val: unknown) => val !== null && typeof val === 'object' && !Array.isArray(val),
  Message: {
    error: vi.fn(),
    success: vi.fn(),
    loading: vi.fn(() => ({ close: vi.fn() })),
  },
}))

import { instance, clearPendingRequests } from './index'

describe('API instance', () => {
  it('is an axios instance with expected defaults', () => {
    expect(instance.defaults.timeout).toBe(600000)
    expect(instance.defaults.headers['Content-Type']).toBe('application/json')
  })

  it('has request interceptors registered', () => {
    // Axios stores interceptors internally
    const reqInterceptors = (instance.interceptors.request as any).handlers
    expect(reqInterceptors.length).toBeGreaterThanOrEqual(3)
  })

  it('has response interceptors registered', () => {
    const resInterceptors = (instance.interceptors.response as any).handlers
    expect(resInterceptors.length).toBeGreaterThanOrEqual(1)
  })
})

describe('clearPendingRequests', () => {
  it('is a function', () => {
    expect(typeof clearPendingRequests).toBe('function')
  })

  it('does not throw when called with no pending requests', () => {
    expect(() => clearPendingRequests()).not.toThrow()
  })
})
