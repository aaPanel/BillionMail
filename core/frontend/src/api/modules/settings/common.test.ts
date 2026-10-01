import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()

vi.mock('@/api', () => ({
  instance: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
  },
}))

vi.mock('@/i18n', () => ({
  default: {
    global: { t: (key: string) => key },
  },
}))

import {
  getSystemConfig,
  getTimezoneList,
  setSystemConfigKey,
  setSslConfig,
  applyCert,
  addIpWhitelist,
  deleteIpWhitelist,
  clearIpWhitelist,
  setReverseProxyDomain,
  clearReverseProxyDomain,
  setApiDocEnabled,
  setBlacklistAutoScan,
  setBlacklistAlert,
  setBlacklistAlertSettings,
} from './common'

describe('settings/common API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getSystemConfig calls GET /settings/get_system_config', () => {
    getSystemConfig()
    expect(mockGet).toHaveBeenCalledWith('/settings/get_system_config')
  })

  it('getTimezoneList calls GET /settings/get_timezone_list', () => {
    getTimezoneList()
    expect(mockGet).toHaveBeenCalledWith('/settings/get_timezone_list')
  })

  it('setSystemConfigKey calls POST /settings/set_system_config_key', () => {
    setSystemConfigKey({ key: 'timezone', value: 'UTC' })
    expect(mockPost).toHaveBeenCalledWith('/settings/set_system_config_key', { key: 'timezone', value: 'UTC' }, expect.any(Object))
  })

  it('setSslConfig calls POST /settings/set_ssl_config', () => {
    setSslConfig({ certPem: 'cert', privateKey: 'key' })
    expect(mockPost).toHaveBeenCalledWith('/settings/set_ssl_config', { certPem: 'cert', privateKey: 'key' }, expect.any(Object))
  })

  it('applyCert calls POST /ssl/console_apply_cert with empty body', () => {
    applyCert()
    expect(mockPost).toHaveBeenCalledWith('/ssl/console_apply_cert', {}, expect.any(Object))
  })

  it('addIpWhitelist calls POST /settings/add_ip_whitelist', () => {
    addIpWhitelist({ ip: '192.168.1.1' })
    expect(mockPost).toHaveBeenCalledWith('/settings/add_ip_whitelist', { ip: '192.168.1.1' }, expect.any(Object))
  })

  it('deleteIpWhitelist calls POST /settings/delete_ip_whitelist', () => {
    deleteIpWhitelist({ id: 10 })
    expect(mockPost).toHaveBeenCalledWith('/settings/delete_ip_whitelist', { id: 10 }, expect.any(Object))
  })

  it('clearIpWhitelist calls POST /settings/set_ip_whitelist with empty list', () => {
    clearIpWhitelist()
    expect(mockPost).toHaveBeenCalledWith('/settings/set_ip_whitelist', { ip_list: [] }, expect.any(Object))
  })

  it('setReverseProxyDomain calls POST /settings/set_reverse_proxy_domain', () => {
    setReverseProxyDomain({ domain: 'proxy.test.com' })
    expect(mockPost).toHaveBeenCalledWith('/settings/set_reverse_proxy_domain', { domain: 'proxy.test.com' }, expect.any(Object))
  })

  it('clearReverseProxyDomain calls POST /settings/delete_reverse_proxy_domain', () => {
    clearReverseProxyDomain()
    expect(mockPost).toHaveBeenCalledWith('/settings/delete_reverse_proxy_domain', {}, expect.any(Object))
  })

  it('setApiDocEnabled calls POST /settings/set_api_doc_swagger', () => {
    setApiDocEnabled({ api_doc_enabled: true })
    expect(mockPost).toHaveBeenCalledWith('/settings/set_api_doc_swagger', { api_doc_enabled: true }, expect.any(Object))
  })

  it('setBlacklistAutoScan calls POST /settings/set_blacklist_auto_scan', () => {
    setBlacklistAutoScan({ enabled: true })
    expect(mockPost).toHaveBeenCalledWith('/settings/set_blacklist_auto_scan', { enabled: true }, expect.any(Object))
  })

  it('setBlacklistAlert calls POST /settings/set_blacklist_alert', () => {
    setBlacklistAlert({ enabled: false })
    expect(mockPost).toHaveBeenCalledWith('/settings/set_blacklist_alert', { enabled: false }, expect.any(Object))
  })

  it('setBlacklistAlertSettings calls POST /settings/set_blacklist_alert_settings', () => {
    const params = {
      name: 'Alert',
      sender_email: 'alert@test.com',
      smtp_password: 'pass',
      smtp_server: 'smtp.test.com',
      smtp_port: 587,
      recipient_list: ['admin@test.com'],
    }
    setBlacklistAlertSettings(params)
    expect(mockPost).toHaveBeenCalledWith('/settings/set_blacklist_alert_settings', params, expect.any(Object))
  })
})
