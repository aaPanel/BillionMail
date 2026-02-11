import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import useInstanceStore, { normalizeUrl } from './instance'

describe('InstanceStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with empty instances', () => {
    const store = useInstanceStore()
    expect(store.instances).toEqual([])
  })

  it('currentInstance is undefined when no match', () => {
    const store = useInstanceStore()
    expect(store.currentInstance).toBeUndefined()
  })

  describe('addInstance', () => {
    it('adds an instance with normalized url', () => {
      const store = useInstanceStore()
      store.addInstance('Prod', 'https://mail.example.com/path')
      expect(store.instances).toHaveLength(1)
      expect(store.instances[0].name).toBe('Prod')
      expect(store.instances[0].url).toBe('https://mail.example.com')
      expect(store.instances[0].id).toBeTruthy()
    })

    it('adds multiple instances', () => {
      const store = useInstanceStore()
      store.addInstance('Prod', 'https://prod.example.com')
      store.addInstance('Staging', 'https://staging.example.com')
      expect(store.instances).toHaveLength(2)
    })

    it('generates unique ids', () => {
      const store = useInstanceStore()
      store.addInstance('A', 'https://a.com')
      store.addInstance('B', 'https://b.com')
      expect(store.instances[0].id).not.toBe(store.instances[1].id)
    })
  })

  describe('updateInstance', () => {
    it('updates name and url', () => {
      const store = useInstanceStore()
      store.addInstance('Old', 'https://old.com')
      const id = store.instances[0].id
      store.updateInstance(id, 'New', 'https://new.com/path')
      expect(store.instances[0].name).toBe('New')
      expect(store.instances[0].url).toBe('https://new.com')
    })

    it('does nothing for unknown id', () => {
      const store = useInstanceStore()
      store.addInstance('Test', 'https://test.com')
      store.updateInstance('nonexistent', 'X', 'https://x.com')
      expect(store.instances).toHaveLength(1)
      expect(store.instances[0].name).toBe('Test')
    })
  })

  describe('removeInstance', () => {
    it('removes by id', () => {
      const store = useInstanceStore()
      store.addInstance('A', 'https://a.com')
      store.addInstance('B', 'https://b.com')
      const id = store.instances[0].id
      store.removeInstance(id)
      expect(store.instances).toHaveLength(1)
      expect(store.instances[0].name).toBe('B')
    })

    it('does nothing for unknown id', () => {
      const store = useInstanceStore()
      store.addInstance('A', 'https://a.com')
      store.removeInstance('nonexistent')
      expect(store.instances).toHaveLength(1)
    })
  })

  describe('currentInstance', () => {
    it('matches instance with same origin as window.location', () => {
      const store = useInstanceStore()
      // happy-dom location.origin is 'http://localhost'
      store.addInstance('Local', window.location.origin)
      expect(store.currentInstance).toBeDefined()
      expect(store.currentInstance!.name).toBe('Local')
    })

    it('does not match different origins', () => {
      const store = useInstanceStore()
      store.addInstance('Remote', 'https://remote.example.com')
      expect(store.currentInstance).toBeUndefined()
    })
  })
})

describe('normalizeUrl', () => {
  it('strips path from url', () => {
    expect(normalizeUrl('https://example.com/path/to/thing')).toBe('https://example.com')
  })

  it('strips trailing slash', () => {
    expect(normalizeUrl('https://example.com/')).toBe('https://example.com')
  })

  it('preserves port', () => {
    expect(normalizeUrl('https://example.com:8443/path')).toBe('https://example.com:8443')
  })

  it('falls back to stripping trailing slash for invalid url', () => {
    expect(normalizeUrl('not-a-url///')).toBe('not-a-url')
  })
})
