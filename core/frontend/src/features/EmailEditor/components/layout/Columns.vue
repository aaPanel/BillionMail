<template>
	<block-section :data="data" block-type="columns" :block-index="index">
		<div ref="columnsLayoutRef" class="columns-layout" :style="columnsStyle">
			<template v-for="(key, idx) in data.children" :key="key">
				<block-cell :blockKey="key" :data="cellMap[key]"> </block-cell>
				<div
					v-if="idx < data.children.length - 1"
					v-show="isBlockSelected"
					class="column-divider"
					:class="{ active: isResizing && currentResizeIndex === idx }"
					@mousedown="startResizing($event, idx)"></div>
			</template>
		</div>
	</block-section>
</template>

<script lang="ts" setup>
import { useStyle } from '../../hooks/useStyle'
import { useConfig } from '../../hooks/useConfig'
import { BlockColumnsType } from '../../types/base'

import BlockSection from './Section.vue'
import BlockCell from './Cell.vue'

const { data } = defineProps({
	index: Number,
	data: {
		type: Object as PropType<BlockColumnsType>,
		required: true,
	},
})

const { columnsConfigMap, cellMap, selectedBlockKey, pageConfig } = useConfig()
const { configToStyle } = useStyle()

// 使用shallowRef代替ref，避免不必要的深层响应式跟踪
const columnsLayoutRef = shallowRef<HTMLElement | null>(null)
const columnWidths = ref<number[]>([])
const isResizing = ref(false)
const currentResizeIndex = ref(-1)
const startX = ref(0)
const startWidths = ref<number[]>([])

// 单元格样式
const columnsStyle = computed(() => {
	const { style } = columnsConfigMap.value[data.key]
	return {
		...configToStyle(style),
		width: pageConfig.value.style.width,
	}
})

// 判断当前区块是否被选中
const isBlockSelected = computed(() => selectedBlockKey.value === data.key)

// 添加一个变量来跟踪更新数据的计时器
const updateDataTimeout = ref<number | null>(null)

// 更新列宽度到数据模型的函数
const updateColumnsDataModel = (leftIndex: number, rightIndex: number) => {
	if (updateDataTimeout.value) {
		clearTimeout(updateDataTimeout.value)
	}

	updateDataTimeout.value = window.setTimeout(() => {
		cellMap.value[data.children[leftIndex]].width = columnWidths.value[leftIndex]
		cellMap.value[data.children[rightIndex]].width = columnWidths.value[rightIndex]
		updateDataTimeout.value = null
	}, 16) // 约一帧的时间
}

// 监听列数变化
watch(
	() => data.children.length,
	() => initColumnWidths()
)

const initColumnWidths = () => {
	if (!data?.children?.length) return

	// 初始化列宽度，确保总和为100%
	const totalColumns = data.children.length
	const defaultWidth = Math.floor(100 / totalColumns)
	const remainder = 100 - defaultWidth * totalColumns

	columnWidths.value = data.children.map((cellKey, index) => {
		const child = cellMap.value[cellKey]
		// 使用有效宽度或计算默认宽度
		if (typeof child.width === 'number' && !isNaN(child.width)) {
			return child.width
		} else {
			// 将余数分配给最后一列
			const width = index === totalColumns - 1 ? defaultWidth + remainder : defaultWidth
			// 更新数据模型
			child.width = width
			return width
		}
	})

	// 验证总宽度
	validateTotalWidth()
}

// 验证并修正总宽度
const validateTotalWidth = () => {
	if (!columnWidths.value || columnWidths.value.length === 0) return

	const total = columnWidths.value.reduce((sum, width) => sum + (isNaN(width) ? 0 : width), 0)
	if (total !== 100) {
		// 如果总宽度不是100%，则调整最后一列的宽度
		const diff = 100 - total
		const lastIndex = columnWidths.value.length - 1
		columnWidths.value[lastIndex] = Math.max(10, (columnWidths.value[lastIndex] || 0) + diff)

		// 同时更新数据模型
		if (data.children[lastIndex]) {
			cellMap.value[data.children[lastIndex]].width = columnWidths.value[lastIndex]
		}
	}
}

const startResizing = (event: MouseEvent, idx: number) => {
	if (idx < 0 || idx >= data.children.length - 1) return

	event.preventDefault() // 阻止默认行为

	isResizing.value = true
	currentResizeIndex.value = idx
	startX.value = event.clientX

	// 确保列宽数组已初始化
	if (columnWidths.value.length === 0) {
		initColumnWidths()
	}

	// 获取当前拖动的两列元素
	const cells = columnsLayoutRef.value?.querySelectorAll('.cell-container')
	if (!cells || cells.length <= idx + 1) {
		isResizing.value = false
		return
	}

	const leftWidth = cells[idx].getBoundingClientRect().width
	const rightWidth = cells[idx + 1].getBoundingClientRect().width

	if (leftWidth > 0 && rightWidth > 0) {
		startWidths.value = [leftWidth, rightWidth]

		// 添加事件监听
		addResizeEventListeners()

		// 禁用选择和设置鼠标样式
		document.body.style.userSelect = 'none'
		document.body.style.cursor = 'col-resize'
	} else {
		isResizing.value = false
	}
}

// 添加调整大小的事件监听器
const addResizeEventListeners = () => {
	// 使用捕获阶段以确保更高的优先级
	document.addEventListener('mousemove', handleMouseMove, { passive: false, capture: true })
	document.addEventListener('mouseup', stopResizing, { passive: false, capture: true })

	// 添加触摸事件支持
	document.addEventListener('touchmove', handleTouchMove, { passive: false, capture: true })
	document.addEventListener('touchend', stopResizing, { passive: false, capture: true })
	document.addEventListener('touchcancel', stopResizing, { passive: false, capture: true })
}

// 移除调整大小的事件监听器
const removeResizeEventListeners = () => {
	document.removeEventListener('mousemove', handleMouseMove, { capture: true })
	document.removeEventListener('mouseup', stopResizing, { capture: true })
	document.removeEventListener('touchmove', handleTouchMove, { capture: true })
	document.removeEventListener('touchend', stopResizing, { capture: true })
	document.removeEventListener('touchcancel', stopResizing, { capture: true })
}

// 处理触摸移动事件
const handleTouchMove = (event: TouchEvent) => {
	if (!isResizing.value || !event.touches[0]) return

	event.preventDefault()

	// 将触摸事件转换为鼠标事件格式
	const touchEvent = {
		clientX: event.touches[0].clientX,
		preventDefault: () => event.preventDefault(),
	} as MouseEvent

	handleMouseMove(touchEvent)
}

// 使用 requestAnimationFrame 优化拖拽操作
let rafId: number | null = null
let lastMoveEvent: MouseEvent | null = null

const handleMouseMove = (event: MouseEvent) => {
	event.preventDefault() // 阻止默认行为

	if (!isResizing.value) return

	// 存储最新的事件对象
	lastMoveEvent = event

	// 如果已经有等待执行的动画帧，则不再请求新的
	if (rafId === null) {
		rafId = requestAnimationFrame(updateColumnWidths)
	}
}

// 使用防抖优化的拖动更新
const updateColumnWidths = () => {
	rafId = null

	if (!lastMoveEvent || !isResizing.value) return

	const event = lastMoveEvent
	const idx = currentResizeIndex.value

	// 确保索引和起始宽度有效
	if (
		idx < 0 ||
		idx >= columnWidths.value.length - 1 ||
		startWidths.value.length < 2 ||
		startWidths.value[0] <= 0 ||
		startWidths.value[1] <= 0
	) {
		return
	}

	// 计算新宽度 - 使用更高效的计算方式
	const totalWidth = startWidths.value[0] + startWidths.value[1]
	const deltaX = event.clientX - startX.value
	const newLeftWidth = Math.max(50, Math.min(totalWidth - 50, startWidths.value[0] + deltaX))

	// 获取两列的总百分比宽度
	const twoColumnsTotalPercent = columnWidths.value[idx] + columnWidths.value[idx + 1]

	// 最小百分比
	const minPercentage = 10
	const maxLeftPercent = twoColumnsTotalPercent - minPercentage

	// 计算新的百分比，确保在可接受范围内
	let newLeftPercent = Math.round((newLeftWidth / totalWidth) * twoColumnsTotalPercent)
	newLeftPercent = Math.max(minPercentage, Math.min(maxLeftPercent, newLeftPercent))
	let newRightPercent = twoColumnsTotalPercent - newLeftPercent

	// 更新宽度
	columnWidths.value[idx] = newLeftPercent
	columnWidths.value[idx + 1] = newRightPercent

	// 更新数据模型
	updateColumnsDataModel(idx, idx + 1)
}

const stopResizing = (event: MouseEvent | TouchEvent) => {
	event.preventDefault()

	isResizing.value = false

	// 取消任何挂起的动画帧
	if (rafId !== null) {
		cancelAnimationFrame(rafId)
		rafId = null
	}

	// 确保最终状态被更新到数据模型
	if (updateDataTimeout.value) {
		clearTimeout(updateDataTimeout.value)
		updateDataTimeout.value = null

		const idx = currentResizeIndex.value
		if (idx >= 0 && idx < columnWidths.value.length - 1) {
			cellMap.value[data.children[idx]].width = columnWidths.value[idx]
			cellMap.value[data.children[idx + 1]].width = columnWidths.value[idx + 1]
		}
	}

	// 移除事件监听器
	removeResizeEventListeners()

	// 恢复正常状态
	document.body.style.userSelect = ''
	document.body.style.cursor = ''

	lastMoveEvent = null
}

onMounted(() => {
	initColumnWidths()
})

onUnmounted(() => {
	// 确保在组件销毁时清理所有事件监听器
	if (rafId !== null) {
		cancelAnimationFrame(rafId)
	}
	if (updateDataTimeout.value) {
		clearTimeout(updateDataTimeout.value)
	}
	removeResizeEventListeners()
})
</script>

<style lang="scss" scoped>
.columns-layout {
	position: relative;
	display: flex;
	margin: 0 auto;

	.column-divider {
		position: relative;
		width: 2px;
		&::after {
			content: '';
			display: block;
			position: absolute;
			top: 0;
			left: 50%;
			width: 2px;
			height: 100%;
			transform: translateX(-50%);
			background-color: #0871c2;
			opacity: 0.5;
			transition: all 0.2s ease;
			cursor: col-resize;
		}
		&:hover,
		&.active {
			&::after {
				width: 4px;
				opacity: 1;
			}
		}
	}
}
</style>
