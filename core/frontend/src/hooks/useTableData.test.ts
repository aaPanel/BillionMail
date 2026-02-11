import { describe, it, expect, vi } from 'vitest'

vi.mock('@/utils', () => ({
  isArray: (val: unknown) => Array.isArray(val),
  isObject: (val: unknown) => val !== null && typeof val === 'object' && !Array.isArray(val),
}))

import { useTableData } from './useTableData'

describe('useTableData', () => {
  it('returns expected shape', () => {
    const result = useTableData({
      params: { page: 1, page_size: 10 },
    })

    expect(result).toHaveProperty('loading')
    expect(result).toHaveProperty('tableList')
    expect(result).toHaveProperty('tableTotal')
    expect(result).toHaveProperty('tableParams')
    expect(result).toHaveProperty('getTableData')
  })

  it('has correct initial state', () => {
    const result = useTableData({
      params: { page: 1, page_size: 10 },
    })

    expect(result.loading.value).toBe(false)
    expect(result.tableList.value).toEqual([])
    expect(result.tableTotal.value).toBe(0)
    expect(result.tableParams.value).toEqual({ page: 1, page_size: 10 })
  })

  it('uses custom loading initial value', () => {
    const result = useTableData({
      params: { page: 1, page_size: 10 },
      loading: true,
    })

    expect(result.loading.value).toBe(true)
  })

  describe('getTableData', () => {
    it('fetches data and populates tableList/tableTotal', async () => {
      const mockData = {
        list: [{ id: 1, name: 'Item 1' }, { id: 2, name: 'Item 2' }],
        total: 2,
      }
      const fetchFn = vi.fn().mockResolvedValue(mockData)

      const result = useTableData({
        params: { page: 1, page_size: 10 },
        fetchFn,
      })

      await result.getTableData()

      expect(fetchFn).toHaveBeenCalledWith({ page: 1, page_size: 10 })
      expect(result.tableList.value).toEqual(mockData.list)
      expect(result.tableTotal.value).toBe(2)
      expect(result.loading.value).toBe(false)
    })

    it('resets page to 1 when resetPage=true', async () => {
      const fetchFn = vi.fn().mockResolvedValue({ list: [], total: 0 })

      const params = { page: 3, page_size: 10 }
      const result = useTableData({ params, fetchFn })

      await result.getTableData(true)

      // The original params object should have page set to 1
      expect(params.page).toBe(1)
    })

    it('sets loading=false even if fetchFn throws', async () => {
      const fetchFn = vi.fn().mockRejectedValue(new Error('Network error'))

      const result = useTableData({
        params: { page: 1, page_size: 10 },
        fetchFn,
      })

      await expect(result.getTableData()).rejects.toThrow('Network error')
      expect(result.loading.value).toBe(false)
    })

    it('handles non-array list in response', async () => {
      const fetchFn = vi.fn().mockResolvedValue({ list: 'not-array', total: 5 })

      const result = useTableData({
        params: { page: 1, page_size: 10 },
        fetchFn,
      })

      await result.getTableData()

      expect(result.tableList.value).toEqual([])
    })

    it('handles non-number total in response', async () => {
      const fetchFn = vi.fn().mockResolvedValue({ list: [], total: 'NaN' })

      const result = useTableData({
        params: { page: 1, page_size: 10 },
        fetchFn,
      })

      await result.getTableData()

      expect(result.tableTotal.value).toBe(0)
    })

    it('does nothing when fetchFn is not provided', async () => {
      const result = useTableData({
        params: { page: 1, page_size: 10 },
      })

      await result.getTableData()

      expect(result.tableList.value).toEqual([])
      expect(result.tableTotal.value).toBe(0)
      expect(result.loading.value).toBe(false)
    })

    it('handles null response gracefully', async () => {
      const fetchFn = vi.fn().mockResolvedValue(null)

      const result = useTableData({
        params: { page: 1, page_size: 10 },
        fetchFn,
      })

      await result.getTableData()

      expect(result.tableList.value).toEqual([])
      expect(result.tableTotal.value).toBe(0)
    })
  })
})
