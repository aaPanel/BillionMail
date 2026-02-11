import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './leads'

describe('leads route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/leads')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 6,
      key: 'leads',
      title: 'Leads',
      titleKey: 'layout.menu.leads',
    })
  })

  it('has name LeadsLayout', () => {
    expect(route.name).toBe('LeadsLayout')
  })

  it('has one child named Leads', () => {
    expect(route.children).toHaveLength(1)
    expect(route.children![0].name).toBe('Leads')
    expect(route.children![0].path).toBe('/leads')
  })
})
