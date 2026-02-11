import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './template'

describe('template route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/template')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 3,
      title: 'template',
      key: 'template',
      titleKey: 'layout.menu.template',
    })
  })

  it('has two children', () => {
    expect(route.children).toHaveLength(2)
  })

  it('first child is template list', () => {
    expect(route.children![0].path).toBe('/template')
    expect(route.children![0].name).toBe('template')
  })

  it('second child is ai-template with chatId param', () => {
    expect(route.children![1].path).toBe('ai-template/:chatId')
    expect(route.children![1].name).toBe('ai-template')
  })
})
