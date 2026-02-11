import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()

vi.mock('@/api', () => ({
  instance: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
  },
}))

import { uploadFile, downloadFile, getLanguages, setLanguage } from './public'

describe('public API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('uploadFile calls POST /file/upload with multipart headers', () => {
    const formData = new FormData()
    uploadFile(formData)
    expect(mockPost).toHaveBeenCalledWith('/file/upload', formData, expect.objectContaining({
      headers: { 'Content-Type': 'multipart/form-data' },
    }))
  })

  it('uploadFile passes onUploadProgress callback', () => {
    const formData = new FormData()
    const progressFn = vi.fn()
    uploadFile(formData, progressFn)
    // Verify the config includes onUploadProgress
    const call = mockPost.mock.calls[0]
    expect(call[2]).toHaveProperty('onUploadProgress')
    // Invoke it and check the callback fires
    call[2].onUploadProgress({ loaded: 50, total: 100 })
    expect(progressFn).toHaveBeenCalledWith({ loaded: 50, total: 100 })
  })

  it('downloadFile calls GET /file/download with blob responseType', () => {
    downloadFile({ file_path: '/path/to/file' })
    expect(mockGet).toHaveBeenCalledWith('/file/download', {
      params: { file_path: '/path/to/file' },
      responseType: 'blob',
    })
  })

  it('getLanguages calls GET /languages/get', () => {
    getLanguages()
    expect(mockGet).toHaveBeenCalledWith('/languages/get')
  })

  it('setLanguage calls POST /languages/set', () => {
    setLanguage({ language: 'zh' })
    expect(mockPost).toHaveBeenCalledWith('/languages/set', { language: 'zh' }, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })
})
