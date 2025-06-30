import { DataTableCreateRowKey, DataTableRowKey, DataTableRowData } from 'naive-ui'
import { useDebounceFn } from '@vueuse/core'
import { isArray, isObject, isNumber } from '@/utils'

type TableParams = DataTableRowData

interface UseDataTableOptions {
	params: TableParams
	rowKey?: DataTableCreateRowKey
	fetchFn: (params: any) => Promise<unknown>
	immediate?: boolean
	autoRefresh?: boolean
	refreshInterval?: number
	useParams?: (params: TableParams) => TableParams
}

// 核心状态管理
interface TableState {
	loading: boolean
}

export const useDataTable = <T>(options: UseDataTableOptions) => {
	const {
		fetchFn,
		params,
		rowKey = row => row.id as DataTableRowKey,
		immediate = true,
		autoRefresh = false,
		refreshInterval = 30000,
		useParams,
	} = options

	// 核心状态
	const state = reactive<TableState>({
		loading: immediate,
	})

	const data = ref<T[]>([])
	const total = ref(0)

	const tableParams = ref(params)

	// 自动刷新定时器
	let timer: number | null = null

	const getParams = () => {
		if (useParams) return useParams(tableParams.value)
		return tableParams.value
	}

	// 防抖的数据获取
	const debouncedFetch = useDebounceFn(async (resetPage = false) => {
		if (resetPage) {
			tableParams.value.page = 1
		}

		state.loading = true

		try {
			const params = getParams()
			const response = await fetchFn(params)

			if (isObject<{ list: T[]; total: number }>(response)) {
				data.value = isArray(response.list) ? response.list : []
				total.value = isNumber(response.total) ? response.total : 0
			}
		} finally {
			state.loading = false
		}
	}, 300)

	// 核心方法
	const fetch = (resetPage = false) => debouncedFetch(resetPage)
	const refresh = () => fetch()
	const reset = () => fetch(true)

	// 排序处理
	const changeSort = (field?: string, order?: 'asc' | 'desc') => {
		tableParams.value.sort_field = field
		tableParams.value.sort_order = order
		fetch(true)
	}

	// 筛选处理
	const changeFilter = (filters: Record<string, any>) => {
		Object.assign(tableParams.value, filters)
		fetch(true)
	}

	// 自动刷新
	const startAutoRefresh = () => {
		stopAutoRefresh()
		if (autoRefresh && refreshInterval > 0) {
			timer = setInterval(() => {
				if (!state.loading) refresh()
			}, refreshInterval)
		}
	}

	const stopAutoRefresh = () => {
		if (timer) {
			clearInterval(timer)
			timer = null
		}
	}

	// 计算属性
	const hasData = computed(() => data.value.length > 0)

	// 生命周期
	onMounted(() => {
		if (immediate) fetch()
		if (autoRefresh) startAutoRefresh()
	})

	onUnmounted(() => {
		stopAutoRefresh()
	})

	// 简化的返回值 - 只返回核心功能
	return {
		// 数据状态
		data: readonly(data),
		total: readonly(total),
		loading: computed(() => state.loading),
		hasData,
		tableParams,

		// 核心方法
		fetch,
		resetTable: reset,
		refreshTable: refresh,

		// 操作方法
		changeSort,
		changeFilter,

		// 配置对象 - 直接用于 naive-ui 表格
		tableConfig: computed(() => ({
			rowKey,
			data: data.value,
			loading: state.loading,
		})),

		pageConfig: computed(() => ({
			page: tableParams.value.page,
			pageSize: tableParams.value.page_size,
			itemCount: total.value,
		})),
	}
}
