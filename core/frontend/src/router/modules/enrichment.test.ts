import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './enrichment'

describe('enrichment route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/enrichment')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 7,
      key: 'enrichment',
      title: 'Enrichment',
      titleKey: 'layout.menu.enrichment',
    })
  })

  it('has name EnrichmentLayout', () => {
    expect(route.name).toBe('EnrichmentLayout')
  })

  it('has one child named Enrichment', () => {
    expect(route.children).toHaveLength(1)
    expect(route.children![0].name).toBe('Enrichment')
    expect(route.children![0].path).toBe('/enrichment')
  })
})
