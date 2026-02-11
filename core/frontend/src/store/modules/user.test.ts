import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock dependencies
vi.mock('@/utils', () => ({
  confirm: vi.fn(),
  isObject: (val: unknown) => val !== null && typeof val === 'object' && !Array.isArray(val),
}))

vi.mock('@/api/modules/user', () => ({
  logout: vi.fn().mockResolvedValue({}),
}))

vi.mock('@/i18n', () => ({
  default: {
    global: {
      t: (key: string) => key,
      locale: { value: 'en' },
    },
  },
}))

vi.mock('@/router', () => ({
  default: {
    push: vi.fn(),
  },
}))

import useUserStore from './user'
import { confirm } from '@/utils'

describe('UserStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('has correct initial state', () => {
    const store = useUserStore()
    expect(store.login.token).toBe('')
    expect(store.login.refresh_token).toBe('')
    expect(store.login.ttl).toBe(0)
    expect(store.login.expire).toBe(0)
  })

  describe('isLogin', () => {
    it('returns false when no token', () => {
      const store = useUserStore()
      expect(store.isLogin).toBeFalsy()
    })

    it('returns false when token is set but expired', () => {
      const store = useUserStore()
      store.login.token = 'some-token'
      store.login.expire = Date.now() - 1000 // expired 1 second ago
      expect(store.isLogin).toBeFalsy()
    })

    it('returns true when token is set and not expired', () => {
      const store = useUserStore()
      store.login.token = 'some-token'
      store.login.expire = Date.now() + 60000 // expires in 60 seconds
      expect(store.isLogin).toBeTruthy()
    })
  })

  describe('setLoginInfo', () => {
    it('sets login info correctly', () => {
      const store = useUserStore()
      const now = Date.now()

      store.setLoginInfo({
        token: 'test-token',
        refresh_token: 'test-refresh',
        ttl: 3600,
      })

      expect(store.login.token).toBe('test-token')
      expect(store.login.refresh_token).toBe('test-refresh')
      expect(store.login.ttl).toBe(3600)
      // expire should be ttl * 1000 + Date.now() (approximately)
      expect(store.login.expire).toBeGreaterThanOrEqual(now + 3600 * 1000)
      expect(store.login.expire).toBeLessThanOrEqual(Date.now() + 3600 * 1000)
    })

    it('makes isLogin truthy after setting valid info', () => {
      const store = useUserStore()
      store.setLoginInfo({
        token: 'valid-token',
        refresh_token: 'refresh',
        ttl: 3600,
      })
      expect(store.isLogin).toBeTruthy()
    })
  })

  describe('resetLoginInfo', () => {
    it('clears all login fields', () => {
      const store = useUserStore()
      store.setLoginInfo({
        token: 'token',
        refresh_token: 'refresh',
        ttl: 3600,
      })

      store.resetLoginInfo()

      expect(store.login.token).toBe('')
      expect(store.login.refresh_token).toBe('')
      expect(store.login.ttl).toBe(0)
      expect(store.login.expire).toBe(0)
    })

    it('makes isLogin falsy after reset', () => {
      const store = useUserStore()
      store.setLoginInfo({
        token: 'token',
        refresh_token: 'refresh',
        ttl: 3600,
      })
      store.resetLoginInfo()
      expect(store.isLogin).toBeFalsy()
    })
  })

  describe('logout', () => {
    it('calls confirm with correct title and content keys', () => {
      const store = useUserStore()
      store.logout()
      expect(confirm).toHaveBeenCalledWith(
        expect.objectContaining({
          title: 'user.logout.title',
          content: 'user.logout.content',
        })
      )
    })

    it('passes onConfirm callback to confirm', () => {
      const store = useUserStore()
      store.logout()
      const call = vi.mocked(confirm).mock.calls[0][0]
      expect(call.onConfirm).toBeTypeOf('function')
    })
  })
})
