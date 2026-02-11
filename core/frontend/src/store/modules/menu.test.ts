import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import useMenuStore from './menu'

describe('MenuStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('has correct initial state', () => {
    const store = useMenuStore()
    expect(store.isCollapse).toBe(false)
    expect(store.menuList).toEqual([])
  })

  describe('setCollapse', () => {
    it('toggles isCollapse from false to true', () => {
      const store = useMenuStore()
      store.setCollapse()
      expect(store.isCollapse).toBe(true)
    })

    it('toggles isCollapse from true to false', () => {
      const store = useMenuStore()
      store.setCollapse() // true
      store.setCollapse() // false
      expect(store.isCollapse).toBe(false)
    })

    it('toggles correctly on multiple calls', () => {
      const store = useMenuStore()
      for (let i = 0; i < 5; i++) {
        store.setCollapse()
      }
      // 5 toggles from false = true
      expect(store.isCollapse).toBe(true)
    })
  })

  describe('setMenuList', () => {
    it('sets menu list', () => {
      const store = useMenuStore()
      const menus = [
        { path: '/home', name: 'Home', component: {} as any },
        { path: '/about', name: 'About', component: {} as any },
      ]
      store.setMenuList(menus as any)
      expect(store.menuList).toHaveLength(2)
      expect(store.menuList[0].path).toBe('/home')
    })

    it('replaces existing menu list', () => {
      const store = useMenuStore()
      store.setMenuList([{ path: '/old', name: 'Old' }] as any)
      store.setMenuList([{ path: '/new', name: 'New' }] as any)
      expect(store.menuList).toHaveLength(1)
      expect(store.menuList[0].path).toBe('/new')
    })

    it('accepts empty array', () => {
      const store = useMenuStore()
      store.setMenuList([{ path: '/x' }] as any)
      store.setMenuList([])
      expect(store.menuList).toEqual([])
    })
  })

  describe('closeSidebar', () => {
    it('sets isCollapse to false', () => {
      const store = useMenuStore()
      store.setCollapse() // true
      store.closeSidebar()
      expect(store.isCollapse).toBe(false)
    })

    it('keeps false if already false', () => {
      const store = useMenuStore()
      store.closeSidebar()
      expect(store.isCollapse).toBe(false)
    })
  })
})
