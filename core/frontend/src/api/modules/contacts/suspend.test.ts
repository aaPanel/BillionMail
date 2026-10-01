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
  getSuspendList,
  getAutoScan,
  deleteSuspend,
  getScanLogs,
  scanGroup,
  clearSuspend,
  setAutoScan,
} from './suspend'

describe('contacts/suspend API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getSuspendList calls GET /abnormal_recipient/list', () => {
    const params = { page: 1, page_size: 10 }
    getSuspendList(params as any)
    expect(mockGet).toHaveBeenCalledWith('/abnormal_recipient/list', { params })
  })

  it('getAutoScan calls GET /abnormal_recipient/check_switch', () => {
    getAutoScan()
    expect(mockGet).toHaveBeenCalledWith('/abnormal_recipient/check_switch')
  })

  it('deleteSuspend calls POST /abnormal_recipient/delete', () => {
    deleteSuspend({ id: 5 })
    expect(mockPost).toHaveBeenCalledWith('/abnormal_recipient/delete', { id: 5 }, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('getScanLogs calls POST /abnormal_recipient/get_scan_log', () => {
    getScanLogs()
    expect(mockPost).toHaveBeenCalledWith('/abnormal_recipient/get_scan_log')
  })

  it('scanGroup calls POST /abnormal_recipient/check_group', () => {
    const params = { group_id: 1, oper: 1 }
    scanGroup(params)
    expect(mockPost).toHaveBeenCalledWith('/abnormal_recipient/check_group', params, expect.any(Object))
  })

  it('clearSuspend calls POST /abnormal_recipient/clear_abnormal with empty body', () => {
    clearSuspend()
    expect(mockPost).toHaveBeenCalledWith('/abnormal_recipient/clear_abnormal', {}, expect.any(Object))
  })

  it('setAutoScan calls POST /abnormal_recipient/set_check_switch', () => {
    setAutoScan({ oper: 1 })
    expect(mockPost).toHaveBeenCalledWith('/abnormal_recipient/set_check_switch', { oper: 1 }, expect.any(Object))
  })
})
