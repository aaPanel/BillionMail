import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './api'

describe('api route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/send')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 3,
      key: 'api',
      title: 'Send API',
      titleKey: 'layout.menu.sendApi',
    })
  })

  it('has one child named Api', () => {
    expect(route.children).toHaveLength(1)
    expect(route.children![0].name).toBe('Api')
    expect(route.children![0].path).toBe('/send')
  })
})
