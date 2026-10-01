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
    isSupported: { value: true },
    copied: { value: false },
  }),
}))

import { useCopy } from './useCopy'

describe('useCopy', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns copied ref and copyText function', () => {
    const result = useCopy()
    expect(result).toHaveProperty('copied')
    expect(result).toHaveProperty('copyText')
    expect(typeof result.copyText).toBe('function')
  })

  it('shows error when value is empty', async () => {
    const { copyText } = useCopy()
    await copyText('')
    expect(mockMessageError).toHaveBeenCalledWith('common.useCopy.noText')
    expect(mockCopy).not.toHaveBeenCalled()
  })

  it('calls copy and shows success for non-empty value', async () => {
    const { copyText } = useCopy()
    await copyText('hello world')
    expect(mockCopy).toHaveBeenCalledWith('hello world')
    expect(mockMessageSuccess).toHaveBeenCalledWith('common.useCopy.success')
  })

  it('does not show success when showSuccess is false', async () => {
    const { copyText } = useCopy()
    await copyText('hello', false)
    expect(mockCopy).toHaveBeenCalledWith('hello')
    expect(mockMessageSuccess).not.toHaveBeenCalled()
  })
})
