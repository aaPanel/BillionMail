import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './automation'

describe('automation route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/automation')
  })

  it('has correct meta with hidden true', () => {
    expect(route.meta).toMatchObject({
      sort: 3,
      key: 'automation',
      title: 'Automation',
      icon: 'i-mdi-refresh-auto',
      hidden: true,
    })
  })

  it('has one child named Automation', () => {
    expect(route.children).toHaveLength(1)
    expect(route.children![0].name).toBe('Automation')
    expect(route.children![0].path).toBe('/automation')
  })
})
