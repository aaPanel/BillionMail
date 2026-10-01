import { describe, it, expect, vi } from 'vitest'

// Mock import.meta.env before importing the module
vi.mock('./is', async () => {
  const actual = await vi.importActual('./is')
  return actual
})

import { getByteUnit, capitalizeFirstLetter, isUrl, getRandomPassword } from './base'

describe('getByteUnit', () => {
  it('formats 0 bytes', () => {
    expect(getByteUnit(0)).toBe('0.00 B')
  })

  it('formats bytes below 1KB', () => {
    expect(getByteUnit(512)).toBe('512.00 B')
    expect(getByteUnit(1023)).toBe('1023.00 B')
  })

  it('formats kilobytes', () => {
    expect(getByteUnit(1024)).toBe('1.00 KB')
    expect(getByteUnit(1536)).toBe('1.50 KB')
  })

  it('formats megabytes', () => {
    expect(getByteUnit(1048576)).toBe('1.00 MB')
    expect(getByteUnit(1572864)).toBe('1.50 MB')
  })

  it('formats gigabytes', () => {
    expect(getByteUnit(1073741824)).toBe('1.00 GB')
  })

  it('formats terabytes', () => {
    expect(getByteUnit(1099511627776)).toBe('1.00 TB')
  })

  it('respects fixed parameter', () => {
    expect(getByteUnit(1536, true, 0)).toBe('2 KB')
    expect(getByteUnit(1536, true, 1)).toBe('1.5 KB')
    expect(getByteUnit(1536, true, 3)).toBe('1.500 KB')
  })

  it('hides unit when isUnit=false', () => {
    expect(getByteUnit(1024, false)).toBe('1.00')
    expect(getByteUnit(1048576, false)).toBe('1.00')
  })

  it('stops at endUnit', () => {
    // 1 MB = 1048576 bytes, should show in KB when endUnit is KB
    expect(getByteUnit(1048576, true, 2, 'KB')).toBe('1024.00 KB')
  })

  it('handles default (no args)', () => {
    expect(getByteUnit()).toBe('0.00 B')
  })

  it('returns empty string for non-number that cant be converted', () => {
    expect(getByteUnit('abc' as any)).toBe('')
  })

  it('handles string numbers', () => {
    expect(getByteUnit('1024' as any)).toBe('1.00 KB')
  })
})

describe('capitalizeFirstLetter', () => {
  it('capitalizes single word', () => {
    expect(capitalizeFirstLetter('hello')).toBe('Hello')
  })

  it('capitalizes multiple words', () => {
    expect(capitalizeFirstLetter('hello world')).toBe('Hello World')
  })

  it('lowercases rest of word', () => {
    expect(capitalizeFirstLetter('HELLO WORLD')).toBe('Hello World')
  })

  it('handles empty string', () => {
    expect(capitalizeFirstLetter('')).toBe('')
  })

  it('handles single character', () => {
    expect(capitalizeFirstLetter('a')).toBe('A')
  })

  it('handles mixed case', () => {
    expect(capitalizeFirstLetter('hELLo wORLd')).toBe('Hello World')
  })
})

describe('isUrl', () => {
  it('validates http URLs', () => {
    expect(isUrl('http://example.com')).toBe(true)
    expect(isUrl('http://www.example.com')).toBe(true)
  })

  it('validates https URLs', () => {
    expect(isUrl('https://example.com')).toBe(true)
    expect(isUrl('https://example.com/path')).toBe(true)
    expect(isUrl('https://example.com/path?q=1')).toBe(true)
  })

  it('validates URLs without protocol', () => {
    expect(isUrl('example.com')).toBe(true)
    expect(isUrl('www.example.com')).toBe(true)
  })

  it('validates IP addresses', () => {
    expect(isUrl('http://192.168.1.1')).toBe(true)
    expect(isUrl('192.168.1.1')).toBe(true)
    expect(isUrl('192.168.1.1:8080')).toBe(true)
  })

  it('validates URLs with ports', () => {
    expect(isUrl('http://example.com:3000')).toBe(true)
    expect(isUrl('http://localhost:8080')).toBe(false) // localhost is not matched by the regex
  })

  it('rejects invalid URLs', () => {
    expect(isUrl('')).toBe(false)
    expect(isUrl('not a url')).toBe(false)
    expect(isUrl('ftp://example.com')).toBe(false)
  })
})

describe('getRandomPassword', () => {
  it('generates password of default length 16', () => {
    const pwd = getRandomPassword()
    expect(pwd).toHaveLength(16)
  })

  it('generates password of specified length', () => {
    expect(getRandomPassword(8)).toHaveLength(8)
    expect(getRandomPassword(20)).toHaveLength(20)
    expect(getRandomPassword(32)).toHaveLength(32)
  })

  it('enforces minimum length of 3', () => {
    expect(getRandomPassword(1)).toHaveLength(3)
    expect(getRandomPassword(2)).toHaveLength(3)
  })

  it('contains at least one lowercase letter', () => {
    // Run multiple times for statistical confidence
    for (let i = 0; i < 20; i++) {
      const pwd = getRandomPassword(8)
      expect(pwd).toMatch(/[a-z]/)
    }
  })

  it('contains at least one uppercase letter', () => {
    for (let i = 0; i < 20; i++) {
      const pwd = getRandomPassword(8)
      expect(pwd).toMatch(/[A-Z]/)
    }
  })

  it('contains at least one number', () => {
    for (let i = 0; i < 20; i++) {
      const pwd = getRandomPassword(8)
      expect(pwd).toMatch(/[0-9]/)
    }
  })

  it('only contains alphanumeric characters', () => {
    for (let i = 0; i < 20; i++) {
      const pwd = getRandomPassword(16)
      expect(pwd).toMatch(/^[a-zA-Z0-9]+$/)
    }
  })

  it('generates different passwords on successive calls', () => {
    const passwords = new Set<string>()
    for (let i = 0; i < 10; i++) {
      passwords.add(getRandomPassword(16))
    }
    // With 62^16 possibilities, collisions are astronomically unlikely
    expect(passwords.size).toBeGreaterThan(1)
  })
})
