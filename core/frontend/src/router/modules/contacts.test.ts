import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './contacts'

describe('contacts route', () => {
  it('has correct path and redirect', () => {
    expect(route.path).toBe('/contacts')
    expect(route.redirect).toBe('/contacts/group')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 4,
      key: 'contacts',
      title: 'Contacts',
      titleKey: 'layout.menu.contacts',
    })
  })

  it('has name ContactsLayout', () => {
    expect(route.name).toBe('ContactsLayout')
  })

  it('first child has sub-routes for group, subscribers, suspend, tags', () => {
    const contacts = route.children![0]
    expect(contacts.name).toBe('Contacts')
    const subRoutes = contacts.children!
    expect(subRoutes).toHaveLength(4)

    expect(subRoutes[0].path).toBe('group')
    expect(subRoutes[0].name).toBe('ContactsGroup')

    expect(subRoutes[1].path).toBe('subscribers')
    expect(subRoutes[1].name).toBe('ContactsSubscribers')

    expect(subRoutes[2].path).toBe('suspend')
    expect(subRoutes[2].name).toBe('ContactsSuspend')

    expect(subRoutes[3].path).toBe('tags')
    expect(subRoutes[3].name).toBe('ContactsTags')
  })

  it('has settings route with id param', () => {
    const settings = route.children!.find(c => c.name === 'ContactsSettings')
    expect(settings).toBeDefined()
    expect(settings!.path).toBe('settings/:id')
  })
})
