import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useEmailEditorStore } from './index'

describe('EmailEditorStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('has correct initial state', () => {
    const store = useEmailEditorStore()

    expect(store.version).toBe('1.0')
    expect(store.columnsSource).toEqual([])
    expect(store.columnsMap).toEqual({})
    expect(store.columnsConfigMap).toEqual({})
    expect(store.cellMap).toEqual({})
    expect(store.cellConfigMap).toEqual({})
    expect(store.blockMap).toEqual({})
    expect(store.blockConfigMap).toEqual({})
    expect(store.selectedBlockKey).toBe('')
    expect(store.selectedBlockType).toBe('')
  })

  describe('pageConfig', () => {
    it('has default page config', () => {
      const store = useEmailEditorStore()
      expect(store.pageConfig.meta.version).toBe('1.0')
      expect(store.pageConfig.meta.createdAt).toBe('')
      expect(store.pageConfig.meta.updatedAt).toBe('')
      expect(store.pageConfig.style.backgroundColor).toBe('#ffffff')
      expect(store.pageConfig.style.width).toBe('500px')
      expect(store.pageConfig.style.fontFamily).toBe('PingFang SC, Microsoft YaHei')
    })

    it('allows updating page config', () => {
      const store = useEmailEditorStore()
      store.pageConfig.style.backgroundColor = '#000000'
      expect(store.pageConfig.style.backgroundColor).toBe('#000000')
    })
  })

  describe('columnsSource', () => {
    it('can add column keys', () => {
      const store = useEmailEditorStore()
      store.columnsSource.push('col-1', 'col-2')
      expect(store.columnsSource).toEqual(['col-1', 'col-2'])
    })
  })

  describe('blockMap', () => {
    it('can add blocks', () => {
      const store = useEmailEditorStore()
      store.blockMap['block-1'] = { key: 'block-1', name: 'Text', type: 'text' }
      expect(store.blockMap['block-1'].name).toBe('Text')
      expect(store.blockMap['block-1'].type).toBe('text')
    })
  })

  describe('selectedBlockKey / selectedBlockType', () => {
    it('can select a block', () => {
      const store = useEmailEditorStore()
      store.selectedBlockKey = 'block-1'
      store.selectedBlockType = 'text'
      expect(store.selectedBlockKey).toBe('block-1')
      expect(store.selectedBlockType).toBe('text')
    })

    it('can deselect', () => {
      const store = useEmailEditorStore()
      store.selectedBlockKey = 'block-1'
      store.selectedBlockKey = ''
      expect(store.selectedBlockKey).toBe('')
    })
  })

  describe('saveFn', () => {
    it('has default noop saveFn', async () => {
      const store = useEmailEditorStore()
      const result = await store.saveFn()
      expect(result).toBe(false)
    })

    it('can be replaced with custom save function', async () => {
      const store = useEmailEditorStore()
      store.saveFn = async () => true
      const result = await store.saveFn()
      expect(result).toBe(true)
    })
  })

  describe('cellMap', () => {
    it('can add cells with children', () => {
      const store = useEmailEditorStore()
      store.cellMap['cell-1'] = {
        key: 'cell-1',
        width: 50,
        name: 'Cell',
        type: 'cell',
        children: ['block-1', 'block-2'],
      }
      expect(store.cellMap['cell-1'].width).toBe(50)
      expect(store.cellMap['cell-1'].children).toHaveLength(2)
    })
  })

  describe('columnsMap', () => {
    it('can add columns with children cells', () => {
      const store = useEmailEditorStore()
      store.columnsMap['col-1'] = {
        key: 'col-1',
        type: 'columns',
        name: 'Columns',
        children: ['cell-1', 'cell-2'],
      }
      expect(store.columnsMap['col-1'].children).toEqual(['cell-1', 'cell-2'])
    })
  })
})
