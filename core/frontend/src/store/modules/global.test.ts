import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('@/utils', () => ({
  isObject: (val: unknown) => val !== null && typeof val === 'object' && !Array.isArray(val),
}))

const mockGetLanguages = vi.fn()
const mockSetLanguageApi = vi.fn()

vi.mock('@/api/modules/public', () => ({
  getLanguages: () => mockGetLanguages(),
  setLanguage: (params: any) => mockSetLanguageApi(params),
}))

import useGlobalStore from './global'

describe('GlobalStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('has correct initial state', () => {
    const store = useGlobalStore()
    expect(store.domainSource).toBe('')
    expect(store.lang).toBe('en')
    expect(store.langList).toEqual([])
    expect(store.isCollapse).toBe(false)
    expect(store.temp_subject).toBe('')
  })

  describe('setCollapse', () => {
    it('toggles collapse state', () => {
      const store = useGlobalStore()
      expect(store.isCollapse).toBe(false)
      store.setCollapse()
      expect(store.isCollapse).toBe(true)
      store.setCollapse()
      expect(store.isCollapse).toBe(false)
    })
  })

  describe('getLang', () => {
    it('fetches and sets language data', async () => {
      const store = useGlobalStore()
      mockGetLanguages.mockResolvedValue({
        current_language: 'zh',
        available_languages: [
          { cn: '中文', name: 'zh' },
          { cn: 'English', name: 'en' },
        ],
      })

      await store.getLang()

      expect(store.lang).toBe('zh')
      expect(store.langList).toHaveLength(2)
      expect(store.langList[0].name).toBe('zh')
    })

    it('does not fetch if langList is already populated', async () => {
      const store = useGlobalStore()
      store.langList = [{ cn: 'English', name: 'en' }]

      await store.getLang()

      expect(mockGetLanguages).not.toHaveBeenCalled()
    })

    it('does not update state if response is not an object', async () => {
      const store = useGlobalStore()
      mockGetLanguages.mockResolvedValue(null)

      await store.getLang()

      expect(store.lang).toBe('en')
      expect(store.langList).toEqual([])
    })
  })

  describe('setLang', () => {
    it('calls setLanguage API with language param', async () => {
      const store = useGlobalStore()
      mockSetLanguageApi.mockResolvedValue({})

      await store.setLang('ja')

      expect(mockSetLanguageApi).toHaveBeenCalledWith({ language: 'ja' })
    })
  })
})
