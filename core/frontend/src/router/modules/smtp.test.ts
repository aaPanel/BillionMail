import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './smtp'

describe('smtp route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/smtp')
  })

  it('has correct meta with key smtp', () => {
    expect(route.meta).toMatchObject({
      sort: 8,
      key: 'smtp',
      title: 'SMTP',
      hidden: false,
    })
  })

  it('has one child', () => {
    expect(route.children).toHaveLength(1)
    expect(route.children![0].name).toBe('SMTP')
    expect(route.children![0].path).toBe('/smtp')
  })
})
