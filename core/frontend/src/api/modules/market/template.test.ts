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
  getTemplateList,
  getTemplateAll,
  getTemplateDetails,
  addTemplate,
  updateTemplate,
  duplicateTemplate,
  deleteTemplate,
} from './template'

describe('market/template API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getTemplateList calls GET /email_template/list', () => {
    const params = { page: 1, page_size: 10 }
    getTemplateList(params as any)
    expect(mockGet).toHaveBeenCalledWith('/email_template/list', { params })
  })

  it('getTemplateAll calls GET /email_template/all', () => {
    getTemplateAll()
    expect(mockGet).toHaveBeenCalledWith('/email_template/all')
  })

  it('getTemplateDetails calls GET /email_template/get', () => {
    getTemplateDetails({ id: '5' })
    expect(mockGet).toHaveBeenCalledWith('/email_template/get', { params: { id: '5' } })
  })

  it('addTemplate calls POST /email_template/create', () => {
    const params = {
      temp_name: 'My Template',
      add_type: 1,
      html_content: '<p>Hello</p>',
      drag_data: '{}',
    }
    addTemplate(params)
    expect(mockPost).toHaveBeenCalledWith('/email_template/create', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('updateTemplate calls POST /email_template/update', () => {
    const params = {
      id: 1,
      temp_name: 'Updated',
      add_type: 1,
      html_content: '<p>Updated</p>',
      drag_data: '{}',
    }
    updateTemplate(params)
    expect(mockPost).toHaveBeenCalledWith('/email_template/update', params, expect.any(Object))
  })

  it('duplicateTemplate calls POST /email_template/copy', () => {
    duplicateTemplate({ id: 3 })
    expect(mockPost).toHaveBeenCalledWith('/email_template/copy', { id: 3 }, expect.any(Object))
  })

  it('deleteTemplate calls POST /email_template/delete', () => {
    deleteTemplate({ id: 7 })
    expect(mockPost).toHaveBeenCalledWith('/email_template/delete', { id: 7 }, expect.any(Object))
  })
})
