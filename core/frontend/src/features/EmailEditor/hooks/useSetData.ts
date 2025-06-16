import { storeToRefs } from 'pinia'
import { useEmailEditorStore } from '../store'
import { useVersion } from './useVersion'

export const useSetData = () => {
	const store = useEmailEditorStore()
	const {
		version,
		columnsSource,
		columnsMap,
		columnsConfigMap,
		cellMap,
		cellConfigMap,
		blockMap,
		blockConfigMap,
		pageConfig,
		selectedBlockKey,
		selectedBlockType,
		saveFn,
	} = storeToRefs(store)

	const setData = (data: string) => {
		const { parsedData } = useVersion(data)
		if (parsedData) {
			pageConfig.value = parsedData.pageConfig
			blockMap.value = parsedData.blockMap
			blockConfigMap.value = parsedData.blockConfigMap
			cellMap.value = parsedData.cellMap
			cellConfigMap.value = parsedData.cellConfigMap
			columnsConfigMap.value = parsedData.columnsConfigMap
			columnsMap.value = parsedData.columnsMap
			columnsSource.value = parsedData.columnsSource
		}
	}

	const getData = () => {
		return JSON.stringify({
			version: version.value,
			pageConfig: toRaw(pageConfig.value),
			blockMap: toRaw(blockMap.value),
			blockConfigMap: toRaw(blockConfigMap.value),
			cellMap: toRaw(cellMap.value),
			cellConfigMap: toRaw(cellConfigMap.value),
			columnsConfigMap: toRaw(columnsConfigMap.value),
			columnsMap: toRaw(columnsMap.value),
			columnsSource: toRaw(columnsSource.value),
		})
	}

	const resetData = () => {
		blockMap.value = {}
		blockConfigMap.value = {}
		cellMap.value = {}
		cellConfigMap.value = {}
		columnsConfigMap.value = {}
		columnsMap.value = {}
		columnsSource.value = []
		selectedBlockKey.value = ''
		selectedBlockType.value = ''
	}

	const setSaveFn = (fn: () => unknown) => {
		saveFn.value = fn
	}

	const autoSaveFn = () => {
		let timer: null | number = null
		const fn = () => {
			if (timer) {
				clearTimeout(timer)
			}
			timer = window.setTimeout(() => {
				// 执行自动保存
				saveFn.value()
			}, 300)
		}
		fn()
	}

	return {
		saveFn,
		getData,
		setData,
		resetData,
		setSaveFn,
		autoSaveFn,
	}
}
