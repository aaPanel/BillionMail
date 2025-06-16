<template>
	<div :ref="setContainerRef" class="canvas-area" @click.stop="onSelectEmpty">
		<div
			ref="canvasRef"
			class="canvas-container"
			:style="pageStyle"
			@mousemove="handleMouseMove"
			@mouseleave="handleMouseLeave">
			<!-- 遍历渲染列，插入占位块 -->
			<template v-for="(item, index) in displayItems" :key="item.key || `placeholder-${index}`">
				<!-- 真实的列组件 -->
				<div
					v-if="item.type === 'columns'"
					:ref="(el: HTMLElement) => registerColumnRef(el, index)"
					class="block-columns-wrapper">
					<block-columns :data="item" :index="getOriginalIndex(index)"> </block-columns>
				</div>
				<!-- 占位块 -->
				<div
					v-if="item.type === 'placeholder'"
					class="block-placeholder"
					:class="{ active: dragState.isDragging }">
					<div class="placeholder-content">
						<span>Place the new content here</span>
					</div>
				</div>
			</template>

			<block-copyright></block-copyright>

			<!-- 空状态提示 -->
			<div v-if="columnsMaps.length === 0 && !dragState.isDragging" class="empty-container">
				<div class="empty-message">No content here. Drag content from left.</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useDrop } from 'vue3-dnd'
import { throttle } from 'lodash-es'
import { useConfig } from '../../hooks/useConfig'
import { BlockColumnsType, DropBaseBlock } from '../../types/base'

import BlockColumns from './Columns.vue'
import BlockCopyright from '../elements/Copyright.vue'

// ------------------------
// 状态和引用管理
// ------------------------

// 获取上下文和状态
const { pageConfig, columnsMaps, selectBlock, insertColumnSourceAt, moveColumnSource } = useConfig()

// DOM引用
const containerRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLElement | null>(null)

// 设置容器引用并初始化拖放
const setContainerRef = (el: HTMLElement | null) => {
	if (!el) return undefined

	containerRef.value = el
	drop(el)
}

// 列引用映射
const columnRefsMap = ref(new Map<number, HTMLElement>())

// 列位置缓存类型
interface ColumnPosition {
	index: number // 数据索引
	domIndex: number // DOM顺序索引
	top: number // 顶部位置
	bottom: number // 底部位置
	center: number // 中心位置
	height: number // 高度
	element: HTMLElement // DOM元素引用
}

// 列位置缓存数组
const columnPositions = ref<ColumnPosition[]>([])

// 缓存控制状态
const positionCache = reactive({
	needsUpdate: true, // 标记是否需要更新
	isUpdating: false, // 防止并发更新
	lastUpdateTime: 0, // 最后更新时间
	lastColumnsLength: 0, // 上次列数量
})

// 鼠标位置状态
const mousePosition = ref({
	x: 0,
	y: 0,
})

// 拖拽状态管理
const dragState = reactive({
	isDragging: false, // 是否处于拖拽中
	isProcessingDrop: false, // 是否正在处理放置
	placeholderIndex: -1, // 占位符索引
	highlightIndex: -1, // 高亮索引
	insertPosition: null as 'before' | 'after' | null, // 插入位置
	currentDragType: null as string | null, // 当前拖拽类型
	sortingActive: false, // 排序模式标记
})

// 排序专用状态
const sortingState = reactive({
	isSorting: false, // 是否排序中
	sourceIndex: -1, // 源元素索引
	lastAppliedTarget: -1, // 最后应用的目标位置
	originalOrder: [] as BlockColumnsType[], // 原始顺序
	draggedElement: null as HTMLElement | null, // 被拖拽元素
	dragSourceDisplayIndex: -1, // 源元素显示索引
})

// 操作计数
const operationCounter = ref(0)

// ------------------------
// 计算属性
// ------------------------

// 计算显示项目（包含占位符）
const displayItems = computed(() => {
	// 排序模式直接返回原始列表
	if (sortingState.isSorting) {
		return columnsMaps.value
	}

	// 在非有效拖拽状态下不显示占位符
	if (!dragState.isDragging || dragState.placeholderIndex < 0) {
		return columnsMaps.value
	}

	// 添加占位符
	const result = [...columnsMaps.value]
	const placeholder: BlockColumnsType = {
		key: 'drag-placeholder',
		type: 'placeholder',
		name: 'Placeholder',
		children: [],
	}

	const safeIndex = Math.min(Math.max(0, dragState.placeholderIndex), result.length)
	result.splice(safeIndex, 0, placeholder)

	return result
})

// 页面样式
const pageStyle = computed(() => {
	return {
		backgroundColor: pageConfig.value.style.backgroundColor,
	}
})

// ------------------------
// 工具函数
// ------------------------

// 获取原始索引
const getOriginalIndex = (displayIndex: number): number => {
	const item = displayItems.value[displayIndex]
	if (!item) return displayIndex

	if (!sortingState.isSorting) return displayIndex

	// 排序模式，查找原始数据索引
	const originalIndex = columnsMaps.value.findIndex(col => col === item)
	return originalIndex === -1 ? displayIndex : originalIndex
}

// 获取显示索引
// const getDisplayIndex = (originalIndex: number): number => {
// 	if (
// 		!dragState.isDragging ||
// 		dragState.placeholderIndex === -1 ||
// 		originalIndex < dragState.placeholderIndex
// 	) {
// 		return originalIndex
// 	}
// 	return originalIndex + 1
// }

// 辅助拖拽工具
const dragUtils = {
	// 检查是否允许当前拖拽类型
	isAllowedDragType(type: string): boolean {
		// 排序操作始终允许
		if (type === 'SORT_COLUMNS_BLOCK') return true

		// 如果没有现有内容，允许所有类型
		if (columnsMaps.value.length === 0) return true

		// 有内容时只允许Columns类型
		return type === 'ADD_COLUMNS_BLOCK'
	},
	// 查找目标位置
	findTargetPosition(
		mouseY: number
	): { targetIndex: number; targetPosition: 'before' | 'after' } | null {
		if (columnPositions.value.length === 0) return null

		// 情况1: 鼠标在第一个元素上方
		if (mouseY < columnPositions.value[0].top) {
			return { targetIndex: 0, targetPosition: 'before' }
		}

		// 情况2: 鼠标在最后一个元素下方
		const lastPos = columnPositions.value[columnPositions.value.length - 1]
		if (mouseY > lastPos.bottom) {
			return { targetIndex: lastPos.index, targetPosition: 'after' }
		}

		// 情况3: 鼠标在元素内部
		for (const pos of columnPositions.value) {
			// 跳过源元素
			if (sortingState.isSorting && pos.index === sortingState.sourceIndex) continue

			if (mouseY >= pos.top && mouseY <= pos.bottom) {
				return {
					targetIndex: pos.index,
					targetPosition: mouseY < pos.center ? 'before' : 'after',
				}
			}
		}

		// 情况4: 鼠标在元素之间
		for (let i = 0; i < columnPositions.value.length - 1; i++) {
			const current = columnPositions.value[i]
			const next = columnPositions.value[i + 1]

			// 跳过与源元素相关的空隙
			if (
				sortingState.isSorting &&
				(current.index === sortingState.sourceIndex || next.index === sortingState.sourceIndex)
			) {
				continue
			}

			if (mouseY > current.bottom && mouseY < next.top) {
				return { targetIndex: current.index, targetPosition: 'after' }
			}
		}

		return null
	},

	// 计算移动目标索引
	getMoveToIndex(
		targetInfo: { targetIndex: number; targetPosition: 'before' | 'after' } | null
	): number {
		if (!targetInfo) return -1
		return targetInfo.targetPosition === 'before'
			? targetInfo.targetIndex
			: targetInfo.targetIndex + 1
	},

	// 判断是否应该跳过移动
	shouldSkipMove(sourceIndex: number, moveToIndex: number, lastAppliedTarget: number): boolean {
		// 移动到自身位置或相邻位置
		if (moveToIndex === sourceIndex || moveToIndex === sourceIndex + 1) {
			return true
		}

		// 重复相同的移动
		if (moveToIndex === lastAppliedTarget) {
			return true
		}

		return false
	},

	// 计算移动后的新源索引
	getNewSourceIndex(sourceIndex: number, moveToIndex: number): number {
		return moveToIndex > sourceIndex ? moveToIndex - 1 : moveToIndex
	},
}

// ------------------------
// 位置计算与更新
// ------------------------

// 注册列引用
const registerColumnRef = (el: HTMLElement | null, index: number) => {
	if (!el) {
		columnRefsMap.value.delete(index)
		return undefined
	}

	// 设置数据索引属性
	const originalIndex = getOriginalIndex(index)
	el.dataset.index = originalIndex.toString()

	// 保存引用
	columnRefsMap.value.set(index, el)
	positionCache.needsUpdate = true
}

// 更新列位置缓存
const updateColumnPositions = (forceSync = false) => {
	// 跳过不必要的更新
	if (positionCache.isUpdating && !forceSync) return
	if (!positionCache.needsUpdate && !forceSync) return

	positionCache.isUpdating = true

	try {
		const container = containerRef.value
		if (!container) return

		const containerRect = container.getBoundingClientRect()
		const positions: ColumnPosition[] = []

		// 直接查询所有列元素
		const elements = container.querySelectorAll('.block-columns-wrapper')

		if (elements.length === 0) {
			columnPositions.value = []
			positionCache.needsUpdate = false
			return
		}

		// 处理每个元素位置
		elements.forEach((el, domIndex) => {
			const rect = el.getBoundingClientRect()
			const dataIndexAttr = (el as HTMLElement).dataset.index
			let dataIndex = dataIndexAttr ? parseInt(dataIndexAttr, 10) : domIndex

			positions.push({
				index: dataIndex,
				domIndex,
				top: rect.top - containerRect.top,
				bottom: rect.bottom - containerRect.top,
				height: rect.height,
				center: rect.top - containerRect.top + rect.height / 2,
				element: el as HTMLElement,
			})
		})

		// 按垂直位置排序
		columnPositions.value = positions.sort((a, b) => a.top - b.top)

		// 更新缓存状态
		positionCache.needsUpdate = false
		positionCache.lastUpdateTime = Date.now()
		positionCache.lastColumnsLength = columnsMaps.value.length
	} finally {
		positionCache.isUpdating = false
	}
}

// 安排位置更新
// const schedulePositionUpdate = () => {
// 	positionCache.needsUpdate = true
// 	requestAnimationFrame(() => {
// 		updateColumnPositions()
// 	})
// }

// 监视列表变化，更新位置
watch(
	() => columnsMaps.value.length,
	(newLength, oldLength) => {
		if (newLength !== oldLength) {
			positionCache.needsUpdate = true

			// 使用nextTick确保DOM已更新
			nextTick(() => {
				updateColumnPositions()
			})
		}
	}
)

const onSelectEmpty = () => {
	selectBlock('', '')
}

// ------------------------
// 拖拽相关处理函数
// ------------------------

// 处理鼠标移动
const handleMouseMove = throttle((event: MouseEvent) => {
	if (dragState.isProcessingDrop) return

	mousePosition.value = {
		x: event.clientX,
		y: event.clientY,
	}

	// 仅在拖拽状态下更新
	if (dragState.isDragging) {
		// 排序模式
		if (sortingState.isSorting) {
			debouncedDragHandler(mousePosition.value)
		}
		// 添加模式
		else {
			updateDropIndicator(mousePosition.value)
		}
	}
}, 16)

// 处理鼠标离开
const handleMouseLeave = () => {
	if (!dragState.isDragging || dragState.isProcessingDrop) return

	resetInsertIndicators()
}

// 重置插入指示器
const resetInsertIndicators = () => {
	dragState.highlightIndex = -1
	dragState.insertPosition = null
}

// 更新添加模式的放置指示器
const updateDropIndicator = (position: { x: number; y: number }) => {
	// 不在拖拽中或排序中则跳过
	if (!dragState.isDragging || dragState.isProcessingDrop || sortingState.isSorting) return

	// 检查是否允许当前拖拽类型
	if (!dragUtils.isAllowedDragType(dragState.currentDragType || '')) {
		// 不允许则重置指示器
		dragState.placeholderIndex = -1
		return
	}

	const container = containerRef.value
	if (!container) return

	const containerRect = container.getBoundingClientRect()
	const mouseY = position.y - containerRect.top

	// 处理空容器
	if (columnPositions.value.length === 0) {
		dragState.placeholderIndex = 0
		return
	}

	// 更新位置缓存
	if (positionCache.needsUpdate) {
		updateColumnPositions()
	}

	// 查找插入位置
	const targetInfo = dragUtils.findTargetPosition(mouseY)
	if (!targetInfo) return

	// 设置占位符和高亮
	const { targetIndex, targetPosition } = targetInfo
	dragState.highlightIndex = targetIndex
	dragState.insertPosition = targetPosition
	dragState.placeholderIndex = targetPosition === 'before' ? targetIndex : targetIndex + 1
}

// 限流后的拖拽处理
const debouncedDragHandler = throttle(
	(position: { x: number; y: number }) => {
		if (!sortingState.isSorting || dragState.isProcessingDrop) return

		// 更新位置缓存
		if (positionCache.needsUpdate) {
			updateColumnPositions(true)
		}

		// 处理排序拖拽
		handleSortingDrag(position)
	},
	16,
	{ leading: true, trailing: true }
)

// 处理排序拖拽
const handleSortingDrag = (position: { x: number; y: number }) => {
	if (!sortingState.isSorting) return

	const containerRect = containerRef.value?.getBoundingClientRect()
	if (!containerRect) return

	const mouseY = position.y - containerRect.top

	// 查找目标位置
	const targetInfo = dragUtils.findTargetPosition(mouseY)
	if (!targetInfo) return

	// 计算移动位置
	const moveToIndex = dragUtils.getMoveToIndex(targetInfo)

	// 判断是否应该跳过
	if (
		dragUtils.shouldSkipMove(sortingState.sourceIndex, moveToIndex, sortingState.lastAppliedTarget)
	) {
		return
	}

	// 更新高亮状态
	dragState.highlightIndex = targetInfo.targetIndex
	dragState.insertPosition = targetInfo.targetPosition

	// 应用排序变更
	operationCounter.value++

	moveColumnSource(sortingState.sourceIndex, moveToIndex)

	// 更新源索引和最后应用的目标
	const newSourceIndex = dragUtils.getNewSourceIndex(sortingState.sourceIndex, moveToIndex)
	sortingState.lastAppliedTarget = moveToIndex
	sortingState.sourceIndex = newSourceIndex

	// 标记需要更新位置
	positionCache.needsUpdate = true

	// 在下一个渲染周期更新位置
	nextTick(() => {
		updateColumnPositions(true)
	})
}

// 开始排序
const startSorting = (sourceIndex: number) => {
	// 防止重复初始化
	if (sortingState.isSorting) return

	// 强制更新位置缓存
	updateColumnPositions(true)

	// 查找源元素显示索引
	let sourceElement = null
	let sourceDisplayIndex = -1

	// 从位置缓存查找
	for (const pos of columnPositions.value) {
		if (pos.index === sourceIndex) {
			sourceDisplayIndex = pos.domIndex
			sourceElement = pos.element
			break
		}
	}

	// 从DOM直接查找
	if (!sourceElement) {
		const container = containerRef.value
		if (container) {
			const elements = container.querySelectorAll('.block-columns-wrapper')
			elements.forEach((el, i) => {
				const dataIndex = parseInt((el as HTMLElement).dataset.index || '-1', 10)
				if (dataIndex === sourceIndex) {
					sourceElement = el as HTMLElement
					sourceDisplayIndex = i
				}
			})
		}
	}

	// 兜底方案
	if (sourceDisplayIndex === -1) {
		sourceDisplayIndex = sourceIndex
	}

	// 初始化排序状态
	sortingState.isSorting = true
	sortingState.sourceIndex = sourceIndex
	sortingState.originalOrder = [...columnsMaps.value]
	sortingState.lastAppliedTarget = -1
	sortingState.draggedElement = sourceElement
	sortingState.dragSourceDisplayIndex = sourceDisplayIndex

	// 更新拖拽状态
	dragState.isDragging = true
	dragState.currentDragType = 'SORT_COLUMNS_BLOCK'
	dragState.sortingActive = true

	// 重置操作计数
	operationCounter.value = 0

	// 更新位置缓存
	positionCache.needsUpdate = true
	updateColumnPositions(true)
}

// 完成排序
const finishSorting = () => {
	// 重置排序状态
	sortingState.isSorting = false
	sortingState.sourceIndex = -1
	sortingState.originalOrder = []
	sortingState.lastAppliedTarget = -1
	sortingState.draggedElement = null
	sortingState.dragSourceDisplayIndex = -1

	// 重置拖拽状态
	dragState.sortingActive = false
	dragState.highlightIndex = -1
	dragState.insertPosition = null

	// 重置操作计数
	operationCounter.value = 0

	// 强制更新位置
	positionCache.needsUpdate = true
	nextTick(() => updateColumnPositions(true))
}

// ------------------------
// 添加组件相关函数
// ------------------------

// 处理组件放置
const handleItemDrop = (item: DropBaseBlock): void => {
	try {
		const { type, blockData } = item

		if (!blockData) return

		// 检查是否允许拖拽类型
		if (!dragUtils.isAllowedDragType(type)) {
			return
		}

		// 根据拖拽类型处理
		if (type === 'ADD_COLUMNS_BLOCK') {
			// 直接添加列组件
			const index =
				dragState.placeholderIndex >= 0 ? dragState.placeholderIndex : columnsMaps.value.length

			insertColumnSourceAt(index)
		} else if (type === 'ADD_BASE_BLOCK') {
			insertColumnSourceAt(0)
		}
	} finally {
		// 强制更新位置缓存
		positionCache.needsUpdate = true
		nextTick(() => updateColumnPositions(true))
	}
}

// 重置拖拽状态
const resetDragStates = () => {
	dragState.isDragging = false
	dragState.isProcessingDrop = false
	dragState.currentDragType = null
	// 确保清除占位符
	dragState.placeholderIndex = -1
	dragState.highlightIndex = -1
	dragState.insertPosition = null
	dragState.sortingActive = false

	// 重置排序状态
	sortingState.isSorting = false
	sortingState.sourceIndex = -1
	sortingState.originalOrder = []
	sortingState.lastAppliedTarget = -1
	sortingState.draggedElement = null
	sortingState.dragSourceDisplayIndex = -1

	// 更新位置缓存
	positionCache.needsUpdate = true
}

// 清理拖拽状态
const clearDragStates = () => {
	resetDragStates()
	nextTick(() => updateColumnPositions())
}

// ------------------------
// 拖拽集成 (vue3-dnd)
// ------------------------

const [, drop] = useDrop<DropBaseBlock, void>(() => ({
	accept: ['ADD_COLUMNS_BLOCK', 'ADD_BASE_BLOCK', 'SORT_COLUMNS_BLOCK'],
	hover: throttle(
		(item, monitor) => {
			// 跳过处理中状态
			if (dragState.isProcessingDrop) return

			// 更精确检测是否真正悬浮在目标上
			const isDirectlyOver = monitor.isOver({ shallow: true })

			// 更新拖拽状态
			dragState.currentDragType = item.type

			// 检查是否允许当前拖拽类型（非排序模式）
			const canDrop = item.type === 'SORT_COLUMNS_BLOCK' || dragUtils.isAllowedDragType(item.type)

			// 只有在直接悬浮且可放置时才显示占位块
			const isDraggingOver = isDirectlyOver && monitor.canDrop() && canDrop

			// 状态变化检测：从悬浮变为非悬浮时清除占位符
			if (dragState.isDragging && !isDraggingOver && !sortingState.isSorting) {
				dragState.placeholderIndex = -1
			}

			// 更新拖拽状态
			dragState.isDragging = isDraggingOver

			// 如果不允许，显示错误并返回
			if (!canDrop && item.type !== 'SORT_COLUMNS_BLOCK') {
				return
			}

			// 如果不在目标上，不处理位置
			if (!isDirectlyOver) {
				return
			}

			const clientOffset = monitor.getClientOffset()
			if (!clientOffset) return

			// 处理不同类型的拖拽
			if (item.type === 'SORT_COLUMNS_BLOCK' && 'sourceIndex' in item) {
				// 排序处理
				if (!sortingState.isSorting) {
					startSorting(item.sourceIndex)
				}

				debouncedDragHandler(clientOffset)
			} else if (item.type === 'ADD_COLUMNS_BLOCK' || item.type === 'ADD_BASE_BLOCK') {
				// 添加组件处理
				updateDropIndicator(clientOffset)
			}
		},
		16,
		{ leading: true, trailing: true }
	),
	drop: (item, monitor) => {
		// 防止重复处理
		if (dragState.isProcessingDrop) return
		dragState.isProcessingDrop = true

		try {
			// 确认是在目标上释放
			if (!monitor.isOver({ shallow: true })) {
				return
			}

			// 根据拖拽类型处理
			if (item.type === 'SORT_COLUMNS_BLOCK') {
				finishSorting()
			} else if (item.type === 'ADD_COLUMNS_BLOCK' || item.type === 'ADD_BASE_BLOCK') {
				handleItemDrop(item)
			}
		} finally {
			// 确保状态重置，使用setTimeout避免状态清理对其他操作的干扰
			setTimeout(() => {
				clearDragStates()
			}, 50)
		}
	},
	// 关键点1: 实时检测拖拽状态
	collect: monitor => {
		// 当拖拽离开区域时，立即触发cleanup
		if (dragState.isDragging && !monitor.isOver()) {
			// 非排序模式下清除占位符
			if (!sortingState.isSorting) {
				dragState.placeholderIndex = -1
			}
			dragState.isDragging = false
		}

		return {}
	},
}))
</script>

<style lang="scss" scoped>
.canvas-area {
	position: relative;
	width: 100%;
	height: 100%;
	background-color: #f4f7fa;
}

.canvas-container {
	position: absolute;
	top: 24px;
	left: 24px;
	right: 24px;
	bottom: 24px;
	box-shadow: 0 4px 12px #ebedf0;
	background-color: #fff;
	overflow-y: auto;
}

.block-columns-wrapper {
	position: relative;
}

.block-placeholder {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 104px;
	border: 2px dashed #1890ff;
	background-color: rgba(24, 144, 255, 0.05);
	border-radius: 4px;
	transition: all 0.2s ease;

	.placeholder-content {
		color: #1890ff;
		font-size: 14px;
		text-align: center;

		span {
			display: block;
		}
	}
}

.empty-container {
	display: flex;
	align-items: center;
	justify-content: center;
	height: calc(100% - 32px);

	.empty-message {
		color: #999;
		font-size: 16px;
		text-align: center;
		padding: 24px;
		border: 2px dashed #e8e8e8;
		border-radius: 4px;
		background-color: #fafafa;
	}
}
</style>
