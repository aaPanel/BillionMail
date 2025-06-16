<template>
	<div class="columns-panel">
		<div class="layout-grid">
			<div v-for="layout in layouts" :key="layout.type" class="layout-option"
				:class="{ active: selectedLayout === layout.type }" @click="selectLayout(layout.type)">
				<div class="layout-preview">
					<div v-for="(column, index) in layout.columns" :key="index" class="column" :style="{ width: column + '%' }">
						<span class="text">{{ column }}</span>
					</div>
				</div>
			</div>
		</div>
		<div v-if="!isStandardLayout" class="layout-option active mt-16px">
			<div class="layout-preview">
				<div v-for="(width, index) in getColumnWidths" :key="index" class="column" :style="{ width: width + '%' }">
					<span class="text">{{ width }}</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useConfig } from '../../../hooks/useConfig'
import { useSetData } from '../../../hooks/useSetData'

const { columnsMap, cellMap, selectedBlockKey, generateNewCell } = useConfig()
const { autoSaveFn } = useSetData()

const columnsConfig = computed(() => {
	return columnsMap.value[selectedBlockKey.value]
})

// 定义布局类型
type LayoutType =
	| 'single'
	| 'two-equal'
	| 'three-equal'
	| 'two-left-small'
	| 'two-right-small'
	| 'four-equal'
	| 'four-left-small'
	| 'four-right-small'

// 定义布局配置
interface LayoutConfig {
	type: LayoutType
	columns: number[]
}



// 布局配置数据
const layouts: LayoutConfig[] = [
	{ type: 'single', columns: [100] },
	{ type: 'two-equal', columns: [50, 50] },
	{ type: 'three-equal', columns: [33, 33, 33] },
	{ type: 'four-equal', columns: [25, 25, 25, 25] },
	{ type: 'two-left-small', columns: [33, 67] },
	{ type: 'two-right-small', columns: [67, 33] },
	{ type: 'four-left-small', columns: [17, 33, 17, 33] },
	{ type: 'four-right-small', columns: [33, 17, 33, 17] },
]

// 创建布局类型到列宽的映射，提高查找效率
const layoutMap = new Map<LayoutType, number[]>(
	layouts.map(layout => [layout.type, layout.columns])
)

// 用于存储用户手动选择的布局
const userSelectedLayout = ref<LayoutType | null>(null)

// 获取列宽数组 - 添加缓存提高性能
const getColumnWidths = computed(() => {
	const children = columnsConfig.value?.children
	if (!children?.length) return []
	return children.map(cellKey => cellMap.value[cellKey].width || 0)
})

// 根据列宽数组找到匹配的布局 - 优化查找算法
const autoDetectedLayout = computed(() => {
	const widths = getColumnWidths.value
	if (!widths.length) return ''

	// 使用for循环替代find，一旦找到匹配就立即返回，避免不必要的迭代
	for (let i = 0; i < layouts.length; i++) {
		const layout = layouts[i]
		if (layout.columns.length !== widths.length) continue

		let isMatch = true
		for (let j = 0; j < widths.length; j++) {
			if (Math.abs(widths[j] - layout.columns[j]) > 3) {
				isMatch = false
				break
			}
		}

		if (isMatch) return layout.type
	}

	return ''
})

// 检查当前布局是否为标准预定义布局
const isStandardLayout = computed((): boolean => {
	return (
		getColumnWidths.value.length > 0 &&
		layouts.some(layout => {
			if (layout.columns.length !== getColumnWidths.value.length) return false

			// 检查每列宽度是否匹配（允许小误差）
			return layout.columns.every((col, idx) => Math.abs(col - getColumnWidths.value[idx]) <= 1)
		})
	)
})

// 最终使用的布局
const selectedLayout = computed(() => {
	// 只有当检测到标准布局时，才返回布局类型
	return isStandardLayout.value ? autoDetectedLayout.value : ''
})

// 根据布局类型获取对应的列宽数组 - 使用Map优化查找
const getLayoutColumns = (layoutType: LayoutType): number[] => {
	return layoutMap.get(layoutType) || [100]
}

// 选择布局并更新 columnsConfig 中的 children
const selectLayout = (layout: LayoutType) => {
	// 更新选中的布局
	userSelectedLayout.value = layout

	// 如果 columnsConfig 不存在，直接返回
	if (!columnsConfig.value) return

	// 获取选中布局的列宽数组
	const columns = getLayoutColumns(layout)
	const currentChildren = columnsConfig.value.children || []

	// 预先计算需要创建多少新cell
	const newCellsCount = Math.max(0, columns.length - currentChildren.length)
	const newCells =
		newCellsCount > 0
			? Array.from({ length: newCellsCount }, () => {
				return generateNewCell()
			})
			: []

	// 创建新的 children 数组，复用现有对象减少GC压力
	const newChildren = columns.map((width, index) => {
		if (index < currentChildren.length) {
			// 直接修改原对象属性而不是创建新对象
			cellMap.value[currentChildren[index]].width = width
			return currentChildren[index]
		}
		// 使用预先创建的新cell
		const newCell = newCells[index - currentChildren.length]
		newCell.width = width
		return newCell.key
	})

	// 批量更新，减少响应式更新次数
	if (selectedBlockKey.value) {
		columnsMap.value[selectedBlockKey.value].children = newChildren
	}

	autoSaveFn()
}
</script>

<style lang="scss" scoped>
.layout-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 16px;
}

.layout-option {
	border: 1px solid #e0e0e0;
	border-radius: 4px;
	cursor: pointer;
	background-color: #f4f4f4;

	&:hover,
	&.active {
		border-color: #333;
		background-color: #fff;

		.column {
			border-left-color: #333;

			.text {
				visibility: visible;
			}
		}
	}
}

.layout-preview {
	display: flex;
	width: 100%;
	height: 36px;
	border-radius: 4px;
	overflow: hidden;
}

.column {
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 12px;
	color: #666;

	&+.column {
		border-left: 1px solid #e0e0e0;
	}

	.text {
		visibility: hidden;
	}
}
</style>
