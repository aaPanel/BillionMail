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
  getVersionInfo,
  getServiceList,
  restartService,
  getServiceConfig,
  saveServiceConfig,
} from './settings'

describe('settings API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getVersionInfo calls GET /settings/get_version', () => {
    getVersionInfo()
    expect(mockGet).toHaveBeenCalledWith('/settings/get_version')
  })

  it('getServiceList calls GET /docker_api/list', () => {
    getServiceList()
    expect(mockGet).toHaveBeenCalledWith('/docker_api/list')
  })

  it('restartService calls POST /docker_api/restart', () => {
    restartService({ container_id: 'abc' })
    expect(mockPost).toHaveBeenCalledWith('/docker_api/restart', { container_id: 'abc' }, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('getServiceConfig calls POST /services/get_config', () => {
    getServiceConfig({ service_type: 'postfix' })
    expect(mockPost).toHaveBeenCalledWith('/services/get_config', { service_type: 'postfix' })
  })

  it('saveServiceConfig calls POST /services/save_config', () => {
    const params = { service_type: 'postfix', content: 'config data' }
    saveServiceConfig(params)
    expect(mockPost).toHaveBeenCalledWith('/services/save_config', params, expect.any(Object))
  })
})
