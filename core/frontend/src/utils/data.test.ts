import { describe, it, expect } from 'vitest'
import { getNumber } from './data'

describe('getNumber', () => {
  it('converts numeric strings to numbers', () => {
    expect(getNumber('42')).toBe(42)
    expect(getNumber('3.14')).toBe(3.14)
    expect(getNumber('0')).toBe(0)
    expect(getNumber('-10')).toBe(-10)
  })

  it('returns numbers as-is', () => {
    expect(getNumber(42)).toBe(42)
    expect(getNumber(0)).toBe(0)
    expect(getNumber(-5)).toBe(-5)
    expect(getNumber(3.14)).toBe(3.14)
  })

  it('returns 0 for non-numeric strings', () => {
    expect(getNumber('abc')).toBe(0)
    expect(getNumber('hello')).toBe(0)
  })

  it('returns 0 for null/undefined', () => {
    expect(getNumber(null)).toBe(0)
    expect(getNumber(undefined)).toBe(0)
  })

  it('returns 0 for NaN', () => {
    expect(getNumber(NaN)).toBe(0)
  })

  it('handles boolean values', () => {
    expect(getNumber(true)).toBe(1)
    expect(getNumber(false)).toBe(0)
  })

  it('handles empty string', () => {
    expect(getNumber('')).toBe(0)
  })

  it('handles whitespace strings', () => {
    expect(getNumber('  42  ')).toBe(42)
  })
})
