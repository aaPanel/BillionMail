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
  getGroupList,
  getGroupAll,
  getGroupInfo,
  createGroup,
  updateGroup,
  deleteGroup,
  getContactCount,
  getContactTagCount,
  exportGroup,
  saveSubscribeSetting,
  saveUnsubscribeSetting,
} from './group'

describe('contacts/group API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getGroupList calls GET /contact/group/list', () => {
    const params = { page: 1, page_size: 10 }
    getGroupList(params as any)
    expect(mockGet).toHaveBeenCalledWith('/contact/group/list', { params })
  })

  it('getGroupAll calls GET /contact/group/all', () => {
    getGroupAll()
    expect(mockGet).toHaveBeenCalledWith('/contact/group/all')
  })

  it('getGroupInfo calls GET /contact/group/info', () => {
    getGroupInfo({ group_id: 5 })
    expect(mockGet).toHaveBeenCalledWith('/contact/group/info', { params: { group_id: 5 } })
  })

  it('createGroup calls POST /contact/group/create', () => {
    const data = {
      create_type: 1,
      name: 'Test Group',
      description: 'desc',
      double_optin: 0,
    }
    createGroup(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/create', data, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('updateGroup calls POST /contact/group/update', () => {
    const data = { group_id: 1, name: 'Updated' }
    updateGroup(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/update', data, expect.any(Object))
  })

  it('deleteGroup calls POST /contact/group/delete', () => {
    const data = { group_ids: [1, 2, 3] }
    deleteGroup(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/delete', data, expect.any(Object))
  })

  it('getContactCount calls POST /contact/group/contact_count', () => {
    const data = { group_ids: [1] }
    getContactCount(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/contact_count', data)
  })

  it('getContactTagCount calls POST /contact/group/tag_contact_count', () => {
    const data = { group_id: 1, tag_ids: [1, 2], tag_logic: 'and' }
    getContactTagCount(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/tag_contact_count', data)
  })

  it('exportGroup calls POST /contact/group/export with blob responseType', () => {
    const data = {
      format: 'csv',
      include_unsubscribe: false,
      group_ids: [1],
      export_type: 1,
    }
    exportGroup(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/export', data, expect.objectContaining({
      responseType: 'blob',
    }))
  })

  it('saveSubscribeSetting calls POST /contact/group/update', () => {
    const data = {
      group_id: 1,
      double_optin: 1,
      send_welcome_email: 1,
      welcome_subject: 'Welcome',
      thank_you_subject: 'Thanks',
      welcome_mail_html: '<p>Hi</p>',
      welcome_mail_drag: '',
      success_url: 'https://example.com',
      confirm_subject: 'Confirm',
      confirm_mail_html: '<p>Confirm</p>',
      confirm_mail_drag: '',
      confirm_url: 'https://example.com/confirm',
      already_url: 'https://example.com/already',
    }
    saveSubscribeSetting(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/update', data, expect.any(Object))
  })

  it('saveUnsubscribeSetting calls POST /contact/group/update_unsubscribe', () => {
    const data = {
      group_id: 1,
      unsubscribe_subject: 'Unsub',
      unsubscribe_redirect_url: 'https://example.com',
      send_unsubscribe_email: 1,
      unsubscribe_mail_html: '<p>Bye</p>',
      unsubscribe_mail_drag: '',
    }
    saveUnsubscribeSetting(data)
    expect(mockPost).toHaveBeenCalledWith('/contact/group/update_unsubscribe', data, expect.any(Object))
  })
})
