import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './market'

describe('market route', () => {
  it('has correct path and redirect', () => {
    expect(route.path).toBe('/market')
    expect(route.redirect).toBe('/market/task')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      key: 'market',
      title: 'Email Marketing',
      titleKey: 'layout.menu.market',
    })
  })

  it('has name MarketLayout', () => {
    expect(route.name).toBe('MarketLayout')
  })

  it('has children with nested routes', () => {
    expect(route.children!.length).toBeGreaterThanOrEqual(3)
  })

  it('first child redirects to /market/task', () => {
    const first = route.children![0]
    expect(first.path).toBe('/market')
    expect(first.redirect).toBe('/market/task')
    expect(first.name).toBe('Market')
  })

  it('first child has task and template sub-routes', () => {
    const market = route.children![0]
    const subRoutes = market.children!
    expect(subRoutes).toHaveLength(2)
    expect(subRoutes[0].path).toBe('task')
    expect(subRoutes[0].name).toBe('MarketTask')
    expect(subRoutes[1].path).toBe('template')
    expect(subRoutes[1].name).toBe('MarketTemplate')
    expect(subRoutes[1].meta?.hidden).toBe(true)
  })

  it('has task edit route', () => {
    const edit = route.children!.find(c => c.name === 'MarketTaskEdit')
    expect(edit).toBeDefined()
    expect(edit!.path).toBe('task/edit')
    expect(edit!.meta?.hidden).toBe(true)
  })

  it('has analytics route with param', () => {
    const analytics = route.children!.find(c => c.name === 'MarketTaskAnalytics')
    expect(analytics).toBeDefined()
    expect(analytics!.path).toBe('task/analytics/:id')
    expect(analytics!.meta?.hidden).toBe(true)
  })
})
