import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockCopy = vi.fn()
const mockMessageError = vi.fn()
const mockMessageSuccess = vi.fn()

vi.mock('@/utils', () => ({
  Message: {
    error: (...args: any[]) => mockMessageError(...args),
    success: (...args: any[]) => mockMessageSuccess(...args),
  },
}))

vi.mock('@/i18n', () => ({
  default: {
    global: { t: (key: string) => key },
  },
}))

vi.mock('@vueuse/core', () => ({
  useClipboard: () => ({
    copy: mockCopy,
    isSupported: { value: false },
    copied: { value: false },
  }),
}))

import { useCopy } from './useCopy'

describe('useCopy (unsupported clipboard)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows failed message when clipboard not supported', async () => {
    const { copyText } = useCopy()
    await copyText('some text')
    expect(mockCopy).not.toHaveBeenCalled()
    expect(mockMessageError).toHaveBeenCalledWith('common.useCopy.failed')
  })
})
