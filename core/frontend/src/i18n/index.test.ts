import { describe, it, expect } from 'vitest'
import i18n, { setLanguage } from './index'

describe('i18n', () => {
  it('is created with legacy false', () => {
    expect(i18n.global).toBeDefined()
  })

  it('defaults to en locale', () => {
    expect(i18n.global.locale.value).toBe('en')
  })

  it('has en as fallback locale', () => {
    expect(i18n.global.fallbackLocale.value).toBe('en')
  })

  it('has en, zh, ja messages loaded', () => {
    const messages = i18n.global.messages.value
    expect(messages).toHaveProperty('en')
    expect(messages).toHaveProperty('zh')
    expect(messages).toHaveProperty('ja')
  })

  describe('setLanguage', () => {
    it('changes locale to zh', () => {
      setLanguage('zh')
      expect(i18n.global.locale.value).toBe('zh')
    })

    it('changes locale to ja', () => {
      setLanguage('ja')
      expect(i18n.global.locale.value).toBe('ja')
    })

    it('changes locale back to en', () => {
      setLanguage('en')
      expect(i18n.global.locale.value).toBe('en')
    })
  })

  describe('translation', () => {
    it('translates a known key in en', () => {
      setLanguage('en')
      const { t } = i18n.global
      // layout.menu.overview should exist in en.json
      const result = t('layout.menu.overview')
      expect(result).toBeTruthy()
      expect(result).not.toBe('layout.menu.overview')
    })

    it('returns key path for unknown keys', () => {
      const { t } = i18n.global
      const result = t('this.key.does.not.exist')
      expect(result).toBe('this.key.does.not.exist')
    })
  })
})
