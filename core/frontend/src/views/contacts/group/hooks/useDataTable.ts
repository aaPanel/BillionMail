import { useDebounceFn } from '@vueuse/core'
import type { DataTableCreateRowKey, DataTableRowKey, DataTableRowData } from 'naive-ui'
import { isArray, isNumber, isObject } from '@/utils'

type TableParams = {
	[key: string]: any
}

interface UseDataTableOptions<T> {
	params: T
	loading?: boolean
	immediate?: boolean
	autoRefresh?: boolean
	refreshInterval?: number
	rowKey?: DataTableCreateRowKey
	fetchFn: (params: TableParams) => Promise<unknown>
	useParams?: (params: T) => TableParams
}

export const useDataTable = <T = DataTableRowData, K = TableParams>(
	options: UseDataTableOptions<K>
) => {
	const {
		params,
		fetchFn,
		useParams,
		loading = true,
		immediate = true,
		rowKey = row => row.id as DataTableRowKey,
	} = options

	const loadingRef = ref(loading)

	const tableData = ref<T[]>([])

	const tableTotal = ref(0)

	const tableParams = ref(params)

	const tableKeys = ref<DataTableRowKey[]>([])

	const getParams = () => {
		if (useParams) return useParams(tableParams.value)
		return tableParams.value
	}

	// 防抖的数据获取
	const debouncedFetch = useDebounceFn(async (resetPage = false) => {
		if (resetPage) {
			tableParams.value.page = 1
		}

		loadingRef.value = true

		try {
			const params = getParams()
			const response = await fetchFn(params)

			if (isObject<{ list: T[]; total: number }>(response)) {
				tableData.value = isArray(response.list) ? response.list : []
				tableTotal.value = isNumber(response.total) ? response.total : 0
			}
		} finally {
			loadingRef.value = false
		}
	}, 300)

	// 核心方法
	const fetchTable = (resetPage = false) => debouncedFetch(resetPage)
	const resetTable = () => fetchTable(true)

	const onUpdateCheckedRowKeys = (keys: DataTableRowKey[]) => {
		tableKeys.value = keys
	}

	const onUpdatePage = (page: number) => {
		tableParams.value.page = page
	}

	const onUpdatePageSize = (pageSize: number) => {
		tableParams.value.page_size = pageSize
	}

	// 生命周期
	onMounted(() => {
		if (immediate) fetchTable()
	})

	return {
		// 数据状态
		tableData,
		tableTotal,
		tableKeys,

		// 核心方法
		fetchTable,
		resetTable,

		// 配置对象 - 直接用于 naive-ui 表格
		tableProps: computed(() => ({
			rowKey,
			data: tableData.value,
			loading: loadingRef.value,
			checkedRowKeys: tableKeys.value,
			onUpdateCheckedRowKeys,
		})),

		pageProps: computed(() => ({
			page: tableParams.value.page,
			pageSize: tableParams.value.page_size,
			itemCount: tableTotal.value,
			onUpdatePage,
			onUpdatePageSize,
		})),
	}
}
