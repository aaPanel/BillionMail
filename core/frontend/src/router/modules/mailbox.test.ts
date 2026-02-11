import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './mailbox'

describe('mailbox route', () => {
  it('has correct path', () => {
    expect(route.path).toBe('/mailbox')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 6,
      key: 'mailbox',
      title: 'MailBoxes',
      titleKey: 'layout.menu.mailboxes',
    })
  })

  it('has one child route named Mailbox', () => {
    expect(route.children).toHaveLength(1)
    expect(route.children![0].name).toBe('Mailbox')
    expect(route.children![0].path).toBe('/mailbox')
  })
})
