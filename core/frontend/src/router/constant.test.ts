import { describe, it, expect } from 'vitest'
import { Layout } from './constant'

describe('router constant', () => {
  it('Layout is a function (lazy import)', () => {
    expect(typeof Layout).toBe('function')
  })

  it('Layout returns a promise', () => {
    const result = Layout()
    expect(result).toBeInstanceOf(Promise)
  })
})
