import { describe, it, expect } from 'vitest'
import { getRandom } from './index'

describe('getRandom', () => {
  it('generates string of default length 12', () => {
    const result = getRandom()
    expect(result).toHaveLength(12)
  })

  it('generates string of specified length', () => {
    expect(getRandom(6)).toHaveLength(6)
    expect(getRandom(8)).toHaveLength(8)
    expect(getRandom(16)).toHaveLength(16)
    expect(getRandom(32)).toHaveLength(32)
  })

  it('only contains hex characters', () => {
    for (let i = 0; i < 20; i++) {
      const result = getRandom(16)
      expect(result).toMatch(/^[0-9a-f]+$/)
    }
  })

  it('generates different values on successive calls', () => {
    const results = new Set<string>()
    for (let i = 0; i < 10; i++) {
      results.add(getRandom(12))
    }
    // Collisions in 12 hex chars are astronomically unlikely
    expect(results.size).toBeGreaterThan(1)
  })

  it('handles length of 1', () => {
    const result = getRandom(1)
    expect(result).toHaveLength(1)
    expect(result).toMatch(/^[0-9a-f]$/)
  })
})
