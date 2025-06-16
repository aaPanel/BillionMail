import { defineStore } from 'pinia'
import {
	ColumnMap,
	ColumnConfigMap,
	CellMap,
	CellConfigMap,
	BlockMap,
	PageConfig,
	BlockConfigMap,
} from '../types/base'

export const useEmailEditorStore = defineStore('email-editor-store', () => {
	const version = ref('1.0')

	// 页面配置
	const pageConfig = ref<PageConfig>({
		meta: {
			version: version.value,
			createdAt: '',
			updatedAt: '',
		},
		style: {
			backgroundColor: '#ffffff',
			width: '500px',
		},
	})

	// columnsSource数据表
	const columnsSource = ref<string[]>([])

	// columnsMap数据表
	const columnsMap = ref<ColumnMap>({})

	// columnsConfigMap数据表
	const columnsConfigMap = ref<ColumnConfigMap>({})

	// cellMap数据表
	const cellMap = ref<CellMap>({})

	// cellConfigMap数据表
	const cellConfigMap = ref<CellConfigMap>({})

	// blockMap数据表
	const blockMap = ref<BlockMap>({})

	// blockMap数据表
	const blockConfigMap = ref<BlockConfigMap>({})

	// 当前选中的块
	const selectedBlockKey = ref('')

	// 当前选中的块类型
	const selectedBlockType = ref('')

	// 自动保存方法
	const saveFn = ref<() => unknown>(async () => false)

	return {
		version,
		pageConfig,
		columnsSource,
		columnsMap,
		columnsConfigMap,
		cellMap,
		cellConfigMap,
		blockMap,
		blockConfigMap,
		selectedBlockKey,
		selectedBlockType,
		saveFn,
	}
})
