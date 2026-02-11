import { describe, it, expect, vi, beforeEach } from 'vitest'
import { nextTick } from 'vue'

vi.mock('@vueuse/core', () => ({
  useDebounceFn: (fn: Function) => fn,
}))

vi.mock('@/utils', () => ({
  isArray: Array.isArray,
  isNumber: (v: unknown) => typeof v === 'number',
  isObject: (v: unknown) => v !== null && typeof v === 'object' && !Array.isArray(v),
}))

import { useDataTable } from './useDataTable'

describe('useDataTable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns expected properties', () => {
    const result = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn: vi.fn(),
    })

    expect(result).toHaveProperty('tableData')
    expect(result).toHaveProperty('tableTotal')
    expect(result).toHaveProperty('tableKeys')
    expect(result).toHaveProperty('tableParams')
    expect(result).toHaveProperty('fetchTable')
    expect(result).toHaveProperty('resetTable')
    expect(result).toHaveProperty('tableProps')
    expect(result).toHaveProperty('batchProps')
    expect(result).toHaveProperty('pageProps')
  })

  it('fetchTable populates tableData and tableTotal', async () => {
    const fetchFn = vi.fn().mockResolvedValue({
      list: [{ id: 1, name: 'Item 1' }, { id: 2, name: 'Item 2' }],
      total: 2,
    })

    const { tableData, tableTotal, fetchTable } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn,
    })

    await fetchTable()
    await nextTick()

    expect(fetchFn).toHaveBeenCalled()
    expect(tableData.value).toHaveLength(2)
    expect(tableTotal.value).toBe(2)
  })

  it('resetTable sets page to 1 before fetching', async () => {
    const fetchFn = vi.fn().mockResolvedValue({ list: [], total: 0 })

    const { tableParams, resetTable } = useDataTable({
      params: { page: 3, page_size: 10 },
      immediate: false,
      fetchFn,
    })

    await resetTable()
    await nextTick()

    expect(tableParams.value.page).toBe(1)
    expect(fetchFn).toHaveBeenCalled()
  })

  it('handles non-object response gracefully', async () => {
    const fetchFn = vi.fn().mockResolvedValue(null)

    const { tableData, tableTotal, fetchTable } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn,
    })

    await fetchTable()
    await nextTick()

    expect(tableData.value).toEqual([])
    expect(tableTotal.value).toBe(0)
  })

  it('clears tableKeys after fetch', async () => {
    const fetchFn = vi.fn().mockResolvedValue({ list: [], total: 0 })

    const { tableKeys, fetchTable } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn,
    })

    tableKeys.value = [1, 2, 3]
    await fetchTable()
    await nextTick()

    expect(tableKeys.value).toEqual([])
  })

  it('uses custom useParams when provided', async () => {
    const fetchFn = vi.fn().mockResolvedValue({ list: [], total: 0 })
    const useParams = vi.fn((params) => ({ ...params, extra: true }))

    const { fetchTable } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn,
      useParams,
    })

    await fetchTable()
    await nextTick()

    expect(useParams).toHaveBeenCalled()
    expect(fetchFn).toHaveBeenCalledWith(expect.objectContaining({ extra: true }))
  })

  it('pageProps returns correct pagination values', () => {
    const { pageProps } = useDataTable({
      params: { page: 2, page_size: 25 },
      immediate: false,
      fetchFn: vi.fn(),
    })

    const props = pageProps.value
    expect(props.page).toBe(2)
    expect(props.pageSize).toBe(25)
    expect(props.itemCount).toBe(0)
    expect(typeof props.onUpdatePage).toBe('function')
    expect(typeof props.onUpdatePageSize).toBe('function')
  })

  it('onUpdatePage updates page in params', () => {
    const { pageProps, tableParams } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn: vi.fn(),
    })

    pageProps.value.onUpdatePage(5)
    expect(tableParams.value.page).toBe(5)
  })

  it('onUpdatePageSize updates page_size in params', () => {
    const { pageProps, tableParams } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn: vi.fn(),
    })

    pageProps.value.onUpdatePageSize(50)
    expect(tableParams.value.page_size).toBe(50)
  })

  it('tableProps includes rowKey and loading state', () => {
    const { tableProps } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      loading: false,
      fetchFn: vi.fn(),
    })

    const props = tableProps.value
    expect(typeof props.rowKey).toBe('function')
    expect(props.loading).toBe(false)
    expect(props.data).toEqual([])
    expect(props.checkedRowKeys).toEqual([])
  })

  it('batchProps has onUpdateCheckedRowKeys handler', () => {
    const { batchProps, tableKeys } = useDataTable({
      params: { page: 1, page_size: 10 },
      immediate: false,
      fetchFn: vi.fn(),
    })

    batchProps.value.onUpdateCheckedRowKeys([1, 2, 3])
    expect(tableKeys.value).toEqual([1, 2, 3])
  })
})
