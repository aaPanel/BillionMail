import { format } from 'date-fns'
import { cloneDeep } from 'lodash-es'
import { storeToRefs } from 'pinia'
import { getRandom } from '../utils'
import { useEmailEditorStore } from '../store'
import { useSetData } from './useSetData'
import { columnsConfig, cellConfig, baseConfigMap } from '../config/config'
import { BlockBaseType, BlockCellType, BlockColumnsType } from '../types/base'

export const useConfig = () => {
	const store = useEmailEditorStore()
	const {
		version,
		selectedBlockKey,
		selectedBlockType,
		columnsSource,
		columnsMap,
		columnsConfigMap,
		cellMap,
		cellConfigMap,
		blockMap,
		blockConfigMap,
		pageConfig,
	} = storeToRefs(store)

	const { autoSaveFn } = useSetData()

	const columnsMaps = computed(() => {
		return columnsSource.value.map(key => {
			return columnsMap.value[key]
		})
	})

	// 选择块
	const selectBlock = (blockKey: string, blockType: string) => {
		selectedBlockKey.value = blockKey
		selectedBlockType.value = blockType
	}

	/**
	 * @description 获取新的列
	 * @returns
	 */
	const getNewColumns = () => {
		const columnKey = getRandom()
		const data: BlockColumnsType = {
			key: columnKey,
			type: 'columns',
			name: 'Columns',
			children: [],
		}
		appendColumnMap(data)
		return data
	}

	/**
	 * @description 添加列配置映射
	 * @param key
	 */
	const appendColumnMap = (data: BlockColumnsType) => {
		const columnsConfig = getNewColumnsConfig(data.key)
		columnsConfigMap.value[data.key] = columnsConfig
	}

	/**
	 * @description 获取新的列配置
	 * @param key
	 * @returns
	 */
	const getNewColumnsConfig = (key: string) => {
		const now = new Date()
		const data = cloneDeep(columnsConfig)
		data.key = key
		data.meta.version = version.value
		data.meta.createdAt = format(now, 'yyyy-MM-dd HH:mm:ss')
		data.meta.updatedAt = format(now, 'yyyy-MM-dd HH:mm:ss')
		return data
	}

	/**
	 * @description 插入列
	 * @param index
	 */
	const insertColumnSourceAt = (index: number) => {
		const newColumn = getNewColumns()
		// 在指定位置插入列
		columnsMap.value[newColumn.key] = newColumn
		insertCellToColumn(newColumn.key)
		columnsSource.value.splice(index, 0, newColumn.key)

		autoSaveFn()
	}

	/**
	 * @description 生成新的单元格
	 * @returns
	 */
	const generateNewCell = () => {
		const cellKey = getRandom()
		const data: BlockCellType = {
			key: cellKey,
			width: 100,
			name: 'Cell',
			type: 'cell',
			children: [],
		}
		cellMap.value[cellKey] = data
		cellConfigMap.value[cellKey] = getNewCellConfig(data.key)
		return data
	}

	/**
	 * @description 获取新的单元格配置
	 * @param key
	 * @returns
	 */
	const getNewCellConfig = (key: string) => {
		const now = new Date()
		const data = cloneDeep(cellConfig)
		data.key = key
		data.meta.version = version.value
		data.meta.createdAt = format(now, 'yyyy-MM-dd HH:mm:ss')
		data.meta.updatedAt = format(now, 'yyyy-MM-dd HH:mm:ss')
		return data
	}

	/**
	 * @description 插入单元格
	 * @param columnKey
	 * @param item
	 */
	const insertCellToColumn = (columnKey: string) => {
		const column = columnsMap.value[columnKey]
		if (column) {
			const newCell = generateNewCell()
			column.children.push(newCell.key)
		}
	}

	// 移动列
	const moveColumnSource = (fromIndex: number, toIndex: number) => {
		// 边界检查
		if (
			fromIndex < 0 ||
			fromIndex >= columnsSource.value.length ||
			toIndex < 0 ||
			toIndex > columnsSource.value.length
		) {
			// console.warn('Move column: 索引超出范围')
			return
		}

		// 取出要移动的列
		const column = columnsSource.value[fromIndex]

		// 创建新数组，避免直接修改原数组
		const newColumns = [...columnsSource.value]

		// 先移除
		newColumns.splice(fromIndex, 1)

		// 计算实际插入位置
		const actualToIndex = fromIndex < toIndex ? toIndex - 1 : toIndex

		// 再插入到新位置
		newColumns.splice(actualToIndex, 0, column)

		// 更新列数组
		columnsSource.value = newColumns

		autoSaveFn()
	}

	/**
	 * @description 删除单元格
	 * @param columnKey
	 * @param cellKey
	 */
	const delColumnsInCell = (columnKey: string, cellKey: string) => {
		const column = columnsMap.value[columnKey]
		if (column) {
			const cellIndex = column.children.findIndex(key => key === cellKey)
			if (cellIndex > -1) {
				column.children.splice(cellIndex, 1)
				delete cellMap.value[cellKey]
			}
		}
	}

	/**
	 * @description 应用新的列宽度到单元格
	 * @param columnKey
	 * @param widths
	 */
	const applyNewWidthColumnsInCell = (columnKey: string, widths: number[]) => {
		const column = columnsMap.value[columnKey]
		if (!column) return

		const columnConfig = columnsMap.value[columnKey]
		if (!columnConfig) return

		// 当前单元格数量
		const currentCellCount = column.children.length
		// 目标单元格数量
		const targetCellCount = widths.length

		// 需要调整的数量差异
		const countDiff = targetCellCount - currentCellCount

		if (countDiff > 0) {
			// 需要添加新单元格
			for (let i = 0; i < countDiff; i++) {
				const newCell = generateNewCell()
				column.children.push(newCell.key)
			}
		} else if (countDiff < 0) {
			// 需要删除多余单元格
			const cellByDel = column.children.slice(targetCellCount, Math.abs(countDiff))
			cellByDel.forEach(cellKey => {
				delColumnsInCell(columnKey, cellKey)
			})
		}

		// 应用新的宽度值到每个单元格
		widths.forEach((width, index) => {
			if (column.children[index]) {
				cellMap.value[column.children[index]].width = width
			}
		})

		autoSaveFn()
	}

	/**
	 * @description 添加块配置映射
	 * @param block
	 */
	const appendBlockConfig = (block: BlockBaseType) => {
		const blockConfig = baseConfigMap.get(block.type)
		if (blockConfig) {
			const now = new Date()
			const data = cloneDeep(blockConfig)
			data.key = block.key
			data.name = block.name
			data.meta.version = version.value
			data.meta.createdAt = format(now, 'yyyy-MM-dd HH:mm:ss')
			data.meta.updatedAt = format(now, 'yyyy-MM-dd HH:mm:ss')
			blockConfigMap.value[block.key] = data
		}
	}

	/**
	 * @description 插入块到单元格
	 * @param index
	 * @param cellKey
	 * @param block
	 */
	const insertBlockToCell = (index: number, cellKey: string, block: BlockBaseType) => {
		const cell = cellMap.value[cellKey]
		if (cell) {
			const cellKey = getRandom()
			const newBlock = {
				key: cellKey,
				type: block.type,
				name: `${block.name}`,
			}
			appendBlockConfig(newBlock)
			blockMap.value[newBlock.key] = newBlock
			cell.children.splice(index, 0, newBlock.key)

			autoSaveFn()
		}
	}

	// 在同一单元格内移动块
	const moveBlockInCell = (cellKey: string, fromIndex: number, toIndex: number) => {
		const cell = cellMap.value[cellKey]
		if (!cell || !cell.children) return

		// 获取要移动的块ID
		const blockId = cell.children[fromIndex]
		if (!blockId) return

		// 如果源索引和目标索引相同，不执行任何操作
		if (fromIndex === toIndex) return

		// 从原位置移除
		cell.children.splice(fromIndex, 1)

		// 计算正确的目标位置（如果目标位置在原位置之后，需要减1，因为已经移除了原位置的元素）
		const targetIndex = fromIndex < toIndex ? toIndex - 1 : toIndex

		// 插入到新位置
		cell.children.splice(targetIndex, 0, blockId)

		autoSaveFn()
	}

	// 在不同单元格之间移动块
	const moveBlockBetweenCells = (
		sourceCellKey: string,
		targetCellKey: string,
		fromIndex: number,
		toIndex: number,
		blockKey: string
	) => {
		// 如果源单元格和目标单元格相同，则调用moveBlockInCell方法
		if (sourceCellKey === targetCellKey) {
			return moveBlockInCell(sourceCellKey, fromIndex, toIndex)
		}

		const sourceCell = cellMap.value[sourceCellKey]
		const targetCell = cellMap.value[targetCellKey]

		if (!sourceCell || !sourceCell.children || !targetCell || !targetCell.children) return

		// 检查源单元格中是否存在该块
		if (sourceCell.children[fromIndex] !== blockKey) return

		// 从源单元格移除
		const block = sourceCell.children.splice(fromIndex, 1)

		// 插入到目标单元格
		targetCell.children.splice(toIndex, 0, block[0])

		autoSaveFn()
	}

	// 复制块
	const duplicateBlock = (cellKey: string, blockKey: string, isAutoSave: boolean = true) => {
		const newBlock = cloneDeep(blockMap.value[blockKey])
		const newBlockConfig = cloneDeep(blockConfigMap.value[blockKey])
		const cell = cellMap.value[cellKey]
		if (cell) {
			const newBlockKey = getRandom()
			newBlock.key = newBlockKey
			newBlockConfig.key = newBlockKey
			newBlockConfig.meta = {
				version: version.value,
				createdAt: format(new Date(), 'yyyy-MM-dd HH:mm:ss'),
				updatedAt: format(new Date(), 'yyyy-MM-dd HH:mm:ss'),
			}
			const blockIndex = cell.children.findIndex(key => key === blockKey)
			blockMap.value[newBlockKey] = newBlock
			blockConfigMap.value[newBlockKey] = newBlockConfig
			if (blockIndex) {
				cell.children.splice(blockIndex + 1, 0, newBlockKey)
			} else {
				cell.children.push(newBlockKey)
			}
		}

		if (isAutoSave) {
			autoSaveFn()
		}
	}

	// 复制单元格以及单元格下的所有块
	const duplicateCell = (columnKey: string, cellKey: string) => {
		const newCell = cloneDeep(cellMap.value[cellKey])
		const newCellConfig = cloneDeep(cellConfigMap.value[cellKey])
		const newCellKey = getRandom()

		newCell.key = newCellKey
		newCell.children = []

		newCellConfig.key = newCellKey
		newCellConfig.meta = {
			version: version.value,
			createdAt: format(new Date(), 'yyyy-MM-dd HH:mm:ss'),
			updatedAt: format(new Date(), 'yyyy-MM-dd HH:mm:ss'),
		}

		cellMap.value[newCellKey] = newCell
		cellConfigMap.value[newCellKey] = newCellConfig

		cellMap.value[cellKey].children.forEach(blockKey => {
			duplicateBlock(newCellKey, blockKey, false)
		})

		const columnsConfig = columnsMap.value[columnKey]
		if (columnsConfig) {
			columnsConfig.children.push(newCellKey)
		}
	}

	// 复制列以及列下的所有块
	const duplicateColumn = (columnKey: string) => {
		const newColumns = cloneDeep(columnsMap.value[columnKey])
		const newColumnsConfig = cloneDeep(columnsConfigMap.value[columnKey])
		const newColumnKey = getRandom()

		newColumns.key = newColumnKey
		newColumns.children = []

		newColumnsConfig.key = newColumnKey
		newColumnsConfig.meta = {
			version: version.value,
			createdAt: format(new Date(), 'yyyy-MM-dd HH:mm:ss'),
			updatedAt: format(new Date(), 'yyyy-MM-dd HH:mm:ss'),
		}
		columnsMap.value[newColumnKey] = newColumns
		columnsConfigMap.value[newColumnKey] = newColumnsConfig

		const columnIndex = columnsSource.value.findIndex(key => key === columnKey)
		columnsSource.value.splice(columnIndex + 1, 0, newColumnKey)
		columnsMap.value[columnKey].children.forEach(cellKey => {
			duplicateCell(newColumnKey, cellKey)
		})

		autoSaveFn()
	}

	// 删除块
	const removeBlock = (cellKey: string, blockKey: string, isAutoSave: boolean = true) => {
		const cell = cellMap.value[cellKey]
		if (cell) {
			const blockIndex = cell.children.findIndex(key => key === blockKey)
			if (blockIndex > -1) {
				selectBlock('', '')
				cell.children.splice(blockIndex, 1)
				delete blockMap.value[blockKey]
				delete blockConfigMap.value[blockKey]
			}
		}
		if (isAutoSave) {
			autoSaveFn()
		}
	}

	// 删除单元格以及单元格下的所有块
	const removeCell = (cellKey: string, columnKey: string) => {
		const cell = cellMap.value[cellKey]
		const columns = columnsMap.value[columnKey]
		if (cell) {
			cell.children.forEach(blockKey => {
				removeBlock(cellKey, blockKey, false)
			})
			const cellIndex = columns.children.findIndex(key => key === cellKey)
			if (cellIndex > -1) {
				delete cellMap.value[cellKey]
				delete cellConfigMap.value[cellKey]
				columns.children.splice(cellIndex, 1)
			}
		}
	}

	// 删除列以及列下的所有块
	const removeColumn = (columnKey: string) => {
		const columnIndex = columnsSource.value.findIndex(key => key === columnKey)
		if (columnIndex > -1) {
			selectBlock('', '')
			columnsMap.value[columnKey].children.forEach(cellKey => {
				removeCell(cellKey, columnKey)
			})
			delete columnsMap.value[columnKey]
			delete columnsConfigMap.value[columnKey]
			columnsSource.value.splice(columnIndex, 1)
		}

		autoSaveFn()
	}

	return {
		columnsSource,
		columnsMap,
		columnsMaps,
		columnsConfigMap,
		cellMap,
		cellConfigMap,
		blockMap,
		blockConfigMap,
		selectedBlockKey,
		selectedBlockType,
		pageConfig,
		selectBlock,
		insertColumnSourceAt,
		moveColumnSource,
		generateNewCell,
		applyNewWidthColumnsInCell,
		delColumnsInCell,
		insertBlockToCell,
		moveBlockInCell,
		moveBlockBetweenCells,
		duplicateBlock,
		duplicateColumn,
		removeBlock,
		removeColumn,
	}
}
