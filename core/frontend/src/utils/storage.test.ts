import { describe, it, expect, beforeEach } from 'vitest'

// storage.ts imports isObject from ./is, which is fine (no env issues)
import {
  getLocalStorage,
  setLocalStorage,
  delLocalStorage,
  setCookie,
  getCookie,
  delCookie,
} from './storage'

describe('localStorage helpers', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  describe('getLocalStorage', () => {
    it('returns default value when key does not exist', () => {
      expect(getLocalStorage('missing', 'default')).toBe('default')
    })

    it('returns stored string value', () => {
      localStorage.setItem('key', 'value')
      expect(getLocalStorage('key')).toBe('value')
    })

    it('returns empty string as default when no defaultVal provided', () => {
      expect(getLocalStorage('missing')).toBe('')
    })

    it('parses JSON when defaultVal is an object', () => {
      const obj = { a: 1, b: 'two' }
      localStorage.setItem('obj', JSON.stringify(obj))
      expect(getLocalStorage('obj', {})).toEqual(obj)
    })

    it('returns default on JSON parse error for object default', () => {
      localStorage.setItem('bad', 'not-json')
      const def = { fallback: true }
      expect(getLocalStorage('bad', def)).toEqual(def)
    })

    it('returns null default when key missing and default is null-like', () => {
      expect(getLocalStorage('missing', 'fallback')).toBe('fallback')
    })
  })

  describe('setLocalStorage', () => {
    it('stores string values', () => {
      setLocalStorage('key', 'value')
      expect(localStorage.getItem('key')).toBe('value')
    })

    it('stores number values as strings', () => {
      setLocalStorage('num', 42)
      expect(localStorage.getItem('num')).toBe('42')
    })

    it('stores objects as JSON', () => {
      const obj = { a: 1 }
      setLocalStorage('obj', obj)
      expect(localStorage.getItem('obj')).toBe(JSON.stringify(obj))
    })

    it('stores boolean as string', () => {
      setLocalStorage('bool', true)
      expect(localStorage.getItem('bool')).toBe('true')
    })
  })

  describe('delLocalStorage', () => {
    it('deletes a single key', () => {
      localStorage.setItem('key', 'val')
      delLocalStorage('key')
      expect(localStorage.getItem('key')).toBeNull()
    })

    it('deletes multiple keys', () => {
      localStorage.setItem('a', '1')
      localStorage.setItem('b', '2')
      localStorage.setItem('c', '3')
      delLocalStorage(['a', 'b'])
      expect(localStorage.getItem('a')).toBeNull()
      expect(localStorage.getItem('b')).toBeNull()
      expect(localStorage.getItem('c')).toBe('3')
    })

    it('does not throw for non-existent key', () => {
      expect(() => delLocalStorage('nonexistent')).not.toThrow()
    })
  })
})

describe('cookie helpers', () => {
  beforeEach(() => {
    // Clear cookies by setting them expired
    document.cookie.split(';').forEach((c) => {
      const name = c.split('=')[0].trim()
      if (name) {
        document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`
      }
    })
  })

  describe('setCookie / getCookie', () => {
    it('sets and gets a cookie', () => {
      setCookie('test', 'value')
      expect(getCookie('test')).toBe('value')
    })

    it('handles special characters via encoding', () => {
      setCookie('special', 'hello world&foo=bar')
      expect(getCookie('special')).toBe('hello world&foo=bar')
    })

    it('returns empty string for non-existent cookie', () => {
      expect(getCookie('nonexistent')).toBe('')
    })
  })

  describe('delCookie', () => {
    it('deletes a cookie by setting it expired', () => {
      setCookie('todel', 'val')
      expect(getCookie('todel')).toBe('val')
      delCookie('todel')
      expect(getCookie('todel')).toBe('')
    })
  })
})
