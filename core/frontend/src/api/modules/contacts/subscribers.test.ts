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
  i18n: {
    global: { t: (key: string) => key },
  },
}))

import {
  getSubscriberList,
  getSubscriberListNdp,
  getSubscriberTrend,
  importSubscribers,
  updateSubscriberGroup,
  editContact,
  editContactNdp,
  deleteSubscriber,
  deleteSubscriberNdp,
  batchSetTag,
} from './subscribers'

describe('contacts/subscribers API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getSubscriberList calls GET /contact/list', () => {
    const params = { page: 1, page_size: 20 }
    getSubscriberList(params as any)
    expect(mockGet).toHaveBeenCalledWith('/contact/list', { params })
  })

  it('getSubscriberListNdp calls GET /contact/list_ndp', () => {
    const params = { page: 1, page_size: 20 }
    getSubscriberListNdp(params as any)
    expect(mockGet).toHaveBeenCalledWith('/contact/list_ndp', { params })
  })

  it('getSubscriberTrend calls GET /contact/trend', () => {
    const params = {
      group_id: 1,
      active: 1,
      last_active_status: 0,
      time_interval: 7,
      tags: '1,2',
    }
    getSubscriberTrend(params)
    expect(mockGet).toHaveBeenCalledWith('/contact/trend', { params })
  })

  it('importSubscribers calls POST /contact/group/import', () => {
    const data = {
      group_ids: [1],
      file_data: 'base64',
      file_type: 'csv',
      contacts: '',
      import_type: 1,
      overwrite: 0,
      default_active: 1,
      status: 1,
    }
    importSubscribers(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/import', data, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('updateSubscriberGroup calls POST /contact/update_group', () => {
    const data = {
      emails: ['a@test.com'],
      active: 1,
      attribs: '{}',
      group_ids: [1, 2],
    }
    updateSubscriberGroup(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/update_group', data, expect.any(Object))
  })

  it('editContact calls POST /contact/edit', () => {
    const data = {
      emails: 'a@test.com',
      active: 1,
      status: 1,
      attribs: '{}',
      group_ids: [1],
    }
    editContact(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/edit', data, expect.any(Object))
  })

  it('editContactNdp calls POST /contact/edit_ndp', () => {
    const data = { id: 1, active: 1, status: 1, attribs: '{}' }
    editContactNdp(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/edit_ndp', data, expect.any(Object))
  })

  it('deleteSubscriber calls POST /contact/delete', () => {
    deleteSubscriber({ emails: ['a@test.com'], status: 1 })
    expect(mockPost).toHaveBeenCalledWith('/contact/delete', { emails: ['a@test.com'], status: 1 }, expect.any(Object))
  })

  it('deleteSubscriberNdp calls POST /contact/delete_ndp', () => {
    deleteSubscriberNdp({ ids: [1, 2] })
    expect(mockPost).toHaveBeenCalledWith('/contact/delete_ndp', { ids: [1, 2] }, expect.any(Object))
  })

  it('batchSetTag calls POST /contact/batch_tags_opt', () => {
    const data = { ids: [1, 2], tag_ids: [10], action: 1 }
    batchSetTag(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/batch_tags_opt', data, expect.any(Object))
  })
})
