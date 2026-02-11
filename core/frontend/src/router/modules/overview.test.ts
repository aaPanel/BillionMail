import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './overview'

describe('overview route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/overview')
  })

  it('has correct meta', () => {
    expect(route.meta).toEqual({
      sort: 0,
      key: 'overview',
      title: 'Overview',
      titleKey: 'layout.menu.overview',
    })
  })

  it('has one child route', () => {
    expect(route.children).toHaveLength(1)
  })

  it('child route points to /overview with name Overview', () => {
    const child = route.children![0]
    expect(child.path).toBe('/overview')
    expect(child.name).toBe('Overview')
  })

  it('child component is a lazy import function', () => {
    const child = route.children![0]
    expect(typeof child.component).toBe('function')
  })
})
