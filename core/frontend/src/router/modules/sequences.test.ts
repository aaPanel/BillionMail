import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './sequences'

describe('sequences route', () => {
  it('has correct path and redirect', () => {
    expect(route.path).toBe('/sequences')
    expect(route.redirect).toBe('/sequences/list')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 5,
      key: 'sequences',
      title: 'Sequences',
      titleKey: 'layout.menu.sequences',
    })
  })

  it('has name SequencesLayout', () => {
    expect(route.name).toBe('SequencesLayout')
  })

  it('first child has list sub-route', () => {
    const sequences = route.children![0]
    expect(sequences.name).toBe('Sequences')
    expect(sequences.redirect).toBe('/sequences/list')
    const subRoutes = sequences.children!
    expect(subRoutes).toHaveLength(1)
    expect(subRoutes[0].path).toBe('list')
    expect(subRoutes[0].name).toBe('SequencesList')
  })

  it('has detail route with id param', () => {
    const detail = route.children!.find(c => c.name === 'SequenceDetail')
    expect(detail).toBeDefined()
    expect(detail!.path).toBe(':id')
  })
})
