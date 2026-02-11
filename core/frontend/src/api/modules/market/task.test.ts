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
  getTaskList,
  getTaskOverview,
  getTaskDetails,
  addTask,
  updateTask,
  deleteTask,
  pauseTask,
  resumeTask,
  sendTestEmail,
  getMailProvider,
  getMailProviderLogs,
} from './task'

describe('market/task API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getTaskList calls GET /batch_mail/task/list', () => {
    const params = { page: 1, page_size: 10 }
    getTaskList(params as any)
    expect(mockGet).toHaveBeenCalledWith('/batch_mail/task/list', { params })
  })

  it('getTaskOverview calls GET /batch_mail/task/stat_chart', () => {
    const params = { task_id: 1, start_time: 100, end_time: 200 }
    getTaskOverview(params)
    expect(mockGet).toHaveBeenCalledWith('/batch_mail/task/stat_chart', { params })
  })

  it('getTaskDetails calls GET /batch_mail/task/find', () => {
    getTaskDetails({ id: 5 })
    expect(mockGet).toHaveBeenCalledWith('/batch_mail/task/find', { params: { id: 5 } })
  })

  it('addTask calls POST /batch_mail/task/create', () => {
    const params = {
      track_open: 1,
      track_click: 1,
      addresser: 'sender@test.com',
      full_name: 'Sender',
      subject: 'Hello',
      group_id: 1,
      template_id: 1,
      is_record: 1,
      warmup: 0,
      unsubscribe: 1,
      threads: 5,
      start_time: 0,
      remark: 'test',
      tag_ids: [],
      tag_logic: 'and',
    }
    addTask(params)
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/task/create', params, expect.objectContaining({
      fetchOptions: expect.objectContaining({ successMessage: true }),
    }))
  })

  it('updateTask calls POST /batch_mail/task/update', () => {
    const params = {
      task_id: 1,
      track_open: 1,
      track_click: 0,
      addresser: 'sender@test.com',
      full_name: 'Sender',
      subject: 'Updated',
      group_id: 2,
      template_id: 3,
      is_record: 1,
      warmup: 0,
      unsubscribe: 0,
      threads: 10,
      start_time: 0,
      remark: '',
      tag_ids: [1],
      tag_logic: 'or',
    }
    updateTask(params)
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/task/update', params, expect.any(Object))
  })

  it('deleteTask calls POST /batch_mail/task/delete', () => {
    deleteTask({ id: 7 })
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/task/delete', { id: 7 }, expect.any(Object))
  })

  it('pauseTask calls POST /batch_mail/task/pause', () => {
    pauseTask({ task_id: 3 })
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/task/pause', { task_id: 3 }, expect.any(Object))
  })

  it('resumeTask calls POST /batch_mail/task/resume', () => {
    resumeTask({ task_id: 3 })
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/task/resume', { task_id: 3 }, expect.any(Object))
  })

  it('sendTestEmail calls POST /batch_mail/task/send_test', () => {
    const params = {
      addresser: 'sender@test.com',
      subject: 'Test',
      recipient: 'user@test.com',
      template_id: 1,
    }
    sendTestEmail(params)
    expect(mockPost).toHaveBeenCalledWith('/batch_mail/task/send_test', params, expect.any(Object))
  })

  it('getMailProvider calls GET /batch_mail/tracking/mail_provider', () => {
    getMailProvider({ task_id: 1, status: 0 })
    expect(mockGet).toHaveBeenCalledWith('/batch_mail/tracking/mail_provider', {
      params: { task_id: 1, status: 0 },
    })
  })

  it('getMailProviderLogs calls GET /batch_mail/tracking/logs', () => {
    const params = { page: 1, page_size: 20, task_id: 1, domain: 'gmail.com' }
    getMailProviderLogs(params)
    expect(mockGet).toHaveBeenCalledWith('/batch_mail/tracking/logs', { params })
  })
})
