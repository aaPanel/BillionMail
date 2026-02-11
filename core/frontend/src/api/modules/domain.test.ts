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
  getDomainList,
  getDomainAll,
  getDomainIpCommand,
  createDomain,
  updateDomain,
  deleteDomain,
  setSsl,
  getSsl,
  freshDnsRecord,
  applyCert,
  setDefaultDomain,
  initAiConfiguration,
  testConnection,
  checkAiConfiguration,
  checkDomainBlacklist,
  getCheckLogs,
} from './domain'

describe('domain API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getDomainList calls GET /domains/list with params', () => {
    const params = { page: 1, page_size: 10, keyword: 'test' }
    getDomainList(params)
    expect(mockGet).toHaveBeenCalledWith('/domains/list', { params })
  })

  it('getDomainAll calls GET /domains/all', () => {
    getDomainAll()
    expect(mockGet).toHaveBeenCalledWith('/domains/all')
  })

  it('getDomainIpCommand calls POST /multi_ip_domain/apply', () => {
    getDomainIpCommand()
    expect(mockPost).toHaveBeenCalledWith('/multi_ip_domain/apply')
  })

  it('createDomain calls POST /domains/create with fetchOptions', () => {
    const params = {
      domain: 'test.com',
      quota: 100,
      mailboxes: 10,
      email: 'a@test.com',
      urls: [],
      hostname: 'mail.test.com',
      outbound_ip: '1.2.3.4',
    }
    createDomain(params)
    expect(mockPost).toHaveBeenCalledWith('/domains/create', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({
        loading: 'domain.api.loading.creating',
        successMessage: true,
      }),
    }))
  })

  it('updateDomain calls POST /domains/update', () => {
    const params = {
      domain: 'test.com',
      quota: 200,
      mailboxes: 20,
      email: 'b@test.com',
      urls: [],
      hostname: 'mail.test.com',
      outbound_ip: '1.2.3.4',
    }
    updateDomain(params)
    expect(mockPost).toHaveBeenCalledWith('/domains/update', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('deleteDomain calls POST /domains/delete', () => {
    deleteDomain({ domain: 'test.com' })
    expect(mockPost).toHaveBeenCalledWith('/domains/delete', { domain: 'test.com' }, expect.any(Object))
  })

  it('setSsl calls POST /domains/set_ssl', () => {
    const params = { domain: 'test.com', certificate: 'cert', key: 'key' }
    setSsl(params)
    expect(mockPost).toHaveBeenCalledWith('/domains/set_ssl', params, expect.any(Object))
  })

  it('getSsl calls GET /domains/get_ssl', () => {
    getSsl({ domain: 'test.com' })
    expect(mockGet).toHaveBeenCalledWith('/domains/get_ssl', { params: { domain: 'test.com' } })
  })

  it('freshDnsRecord calls POST /domains/fresh_dns_records', () => {
    freshDnsRecord({ domain: 'test.com' })
    expect(mockPost).toHaveBeenCalledWith('/domains/fresh_dns_records', { domain: 'test.com' }, expect.any(Object))
  })

  it('applyCert calls POST /ssl/apply_cert', () => {
    applyCert({ domain: 'test.com' })
    expect(mockPost).toHaveBeenCalledWith('/ssl/apply_cert', { domain: 'test.com' }, expect.any(Object))
  })

  it('setDefaultDomain calls POST /domains/set_default_domain', () => {
    setDefaultDomain({ domain: 'test.com' })
    expect(mockPost).toHaveBeenCalledWith('/domains/set_default_domain', { domain: 'test.com' }, expect.any(Object))
  })

  it('initAiConfiguration calls POST /askai/project/create', () => {
    const params = { domain: 'test.com', urls: ['https://example.com'] }
    initAiConfiguration(params)
    expect(mockPost).toHaveBeenCalledWith('/askai/project/create', params, expect.any(Object))
  })

  it('testConnection calls POST /multi_ip_domain/test', () => {
    const params = { domain: 'test.com', outbound_ip: '1.2.3.4' }
    testConnection(params)
    expect(mockPost).toHaveBeenCalledWith('/multi_ip_domain/test', params, expect.any(Object))
  })

  it('checkAiConfiguration calls POST /askai/supplier/status', () => {
    checkAiConfiguration()
    expect(mockPost).toHaveBeenCalledWith('/askai/supplier/status')
  })

  it('checkDomainBlacklist calls POST /domain_blocklist/check', () => {
    checkDomainBlacklist({ a_record: '1.2.3.4' })
    expect(mockPost).toHaveBeenCalledWith('/domain_blocklist/check', { a_record: '1.2.3.4' }, expect.any(Object))
  })

  it('getCheckLogs calls GET /domain_blocklist/logs', () => {
    getCheckLogs({ path: '/some/path' })
    expect(mockGet).toHaveBeenCalledWith('/domain_blocklist/logs', { params: { path: '/some/path' } })
  })
})
