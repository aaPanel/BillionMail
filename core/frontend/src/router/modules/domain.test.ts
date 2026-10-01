import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './domain'

describe('domain route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/domain')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 5,
      key: 'domain',
      title: 'MailDomain',
      titleKey: 'layout.menu.domain',
    })
  })

  it('has two children', () => {
    expect(route.children).toHaveLength(2)
  })

  it('first child is the domain list view', () => {
    const child = route.children![0]
    expect(child.path).toBe('/domain')
    expect(child.name).toBe('Domain')
  })

  it('second child is edit domain with param', () => {
    const child = route.children![1]
    expect(child.path).toBe('edit-domain/:domain')
    expect(child.name).toBe('EditDomain')
  })
})
