import { describe, it, expect, vi } from 'vitest'

vi.mock('@/router/constant', () => ({
  Layout: () => Promise.resolve({ template: '<router-view />' }),
}))

import route from './settings'

describe('settings route', () => {
  it('has correct path and redirect', () => {
    expect(route.path).toBe('/settings')
    expect(route.redirect).toBe('/settings/common')
  })

  it('has correct meta', () => {
    expect(route.meta).toMatchObject({
      sort: 9,
      key: 'settings',
      title: 'Settings',
      titleKey: 'layout.menu.settings',
    })
  })

  it('has name SettingsLayout', () => {
    expect(route.name).toBe('SettingsLayout')
  })

  it('first child has sub-routes for common, service, bcc, forward, ai-model, send-queue', () => {
    const settings = route.children![0]
    expect(settings.name).toBe('Settings')
    expect(settings.redirect).toBe('/settings/common')
    const subRoutes = settings.children!
    expect(subRoutes).toHaveLength(6)

    expect(subRoutes[0].path).toBe('common')
    expect(subRoutes[0].name).toBe('SettingsCommon')

    expect(subRoutes[1].path).toBe('service')
    expect(subRoutes[1].name).toBe('SettingsService')

    expect(subRoutes[2].path).toBe('bcc')
    expect(subRoutes[2].name).toBe('SettingsBcc')

    expect(subRoutes[3].path).toBe('forward')
    expect(subRoutes[3].name).toBe('SettingsForward')

    expect(subRoutes[4].path).toBe('ai-model')
    expect(subRoutes[4].name).toBe('AiModel')

    expect(subRoutes[5].path).toBe('send-queue')
    expect(subRoutes[5].name).toBe('SendQueue')
  })

  it('has rspamd route marked hidden', () => {
    const rspamd = route.children!.find(c => c.name === 'SettingsRspamd')
    expect(rspamd).toBeDefined()
    expect(rspamd!.path).toBe('rspamd')
    expect(rspamd!.meta?.hidden).toBe(true)
  })
})
