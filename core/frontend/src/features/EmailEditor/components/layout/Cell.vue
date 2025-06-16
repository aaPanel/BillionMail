<template>
	<div
		:ref="setContainerRef"
		class="cell-container"
		:class="{ 'drop-active': isOver }"
		:style="cellStyle"
		:data-cell-key="blockKey">
		<!-- 包含占位块的内容区域 -->
		<div ref="contentRef" class="cell-content">
			<!-- 已有组件 -->
			<template v-for="(item, index) in displayItems" :key="item.key || `placeholder-${index}`">
				<!-- 真实的块组件 - 添加包装div以便获取DOM引用 -->
				<div
					v-if="item.type !== 'placeholder'"
					:ref="(el: HTMLElement) => registerBlockRef(el, index)"
					class="block-wrapper">
					<component
						:is="resolveBlockComponent(item.type)"
						:data="item"
						:block-index="getOriginalIndex(index)"
						:cell-key="blockKey"
						class="block-item">
					</component>
				</div>
				<!-- 占位符 -->
				<div v-else class="block-placeholder" :class="{ active: isOver }">
					<div class="placeholder-content">
						<span>Place new content here</span>
					</div>
				</div>
			</template>
		</div>
		<!-- 空状态占位符 -->
		<div v-if="!hasChildren" class="column-placeholder">
			<span>Drag and drop contents to this column</span>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { throttle, debounce } from 'lodash-es'
import { useDrop } from 'vue3-dnd'
import { useStyle } from '../../hooks/useStyle'
import { useConfig } from '../../hooks/useConfig'
import { BlockType, BlockCellType, DropBaseBlock, BlockBaseType } from '../../types/base'

// 导入各种块组件
import BlockButton from '../elements/Button.vue'
import BlockLink from '../elements/Link.vue'
import BlockDivider from '../elements/Divider.vue'
import BlockHeader from '../elements/Header.vue'
import BlockText from '../elements/Text.vue'
import BlockImage from '../elements/Image.vue'
import BlockMenu from '../elements/Menu.vue'
import BlockCopyright from '../elements/Copyright.vue'

const { data, blockKey } = defineProps({
	data: {
		type: Object as PropType<BlockCellType>,
		required: true,
	},
	blockKey: {
		type: String,
		required: true,
	},
})

const {
	cellMap,
	cellConfigMap,
	blockMap,
	insertBlockToCell,
	moveBlockInCell,
	moveBlockBetweenCells,
} = useConfig()

const { configToStyle } = useStyle()

// 使用ref进行DOM引用，避免重复查询
const containerRef = ref<HTMLElement | null>(null)
const contentRef = ref<HTMLElement | null>(null)
const blockRefs = ref<HTMLElement[]>([])

// 设置容器引用
const setContainerRef = (el: HTMLElement | null) => {
	if (!el) return undefined

	containerRef.value = el
	drop(el)
}

// 设置块引用
const registerBlockRef = (el: HTMLElement, index: number) => {
	if (!el) return undefined

	// 确保数组长度足够
	if (blockRefs.value.length <= index) {
		blockRefs.value = blockRefs.value.concat(
			new Array(index - blockRefs.value.length + 1).fill(null)
		)
	}
	blockRefs.value[index] = el
}

// 单元格宽度
const width = computed(() => {
	return `${data.width}%`
})

// 单元格样式
const cellStyle = computed(() => {
	const { style } = cellConfigMap.value[blockKey]
	return {
		...configToStyle(style),
		width: width.value,
	}
})

// 子元素列表
const children = computed(() => {
	if (cellMap.value[blockKey]) {
		return cellMap.value[blockKey].children
			.filter(key => !!blockMap.value[key])
			.map(key => {
				return blockMap.value[key]
			})
	}
	return []
})

// 判断是否有子元素
const hasChildren = computed(() => {
	return children.value && children.value.length > 0
})

// 拖放状态
const isOver = ref(false)
const dropIndex = ref(-1)
const dragSourceIndex = ref(-1)
const dragSourceCell = ref('')
const dragItemType = ref('')

// 记录鼠标当前位置
const mousePosition = ref({
	x: 0,
	y: 0,
})

// 判断是否在同一单元格内拖拽
const isDraggingInSameCell = computed(() => {
	return isOver.value && dragSourceCell.value === blockKey && dragSourceIndex.value !== -1
})

// 块元素位置类型
type BlockPositionType = {
	index: number
	top: number
	bottom: number
	center: number
	height: number // 高度属性
}

// 为块元素位置创建缓存
const blockPositions = ref<BlockPositionType[]>([])

// 获取原始索引（不包含占位块）
const getOriginalIndex = (displayIndex: number) => {
	if (!isOver.value || dropIndex.value === -1 || displayIndex < dropIndex.value) {
		return displayIndex
	}
	return displayIndex - 1
}

// 计算显示项目列表（包含真实块和占位块）
const displayItems = computed(() => {
	if (!hasChildren.value || !isOver.value || dropIndex.value === -1) {
		return children.value || []
	}

	// 创建包含占位块的新数组
	const result = [...(children.value || [])]

	// 如果是拖拽排序，不显示占位符，而是实时移动当前块
	if (isDraggingInSameCell.value) {
		// 避免不必要的操作：如果拖回原位置，直接返回原数组
		if (dragSourceIndex.value === dropIndex.value) {
			return result
		}

		// 从原位置移除
		const draggedItem = result.splice(dragSourceIndex.value, 1)[0]
		// 插入到新位置
		result.splice(
			dropIndex.value > dragSourceIndex.value ? dropIndex.value - 1 : dropIndex.value,
			0,
			draggedItem
		)
		return result
	}

	// 跨单元格拖拽或添加新块时显示占位符
	const placeholder: BlockBaseType = {
		key: 'drag-placeholder',
		type: 'placeholder',
		name: 'Placeholder',
	}

	// 插入占位块到计算出的位置
	if (dropIndex.value <= result.length) {
		result.splice(dropIndex.value, 0, placeholder)
	} else {
		result.push(placeholder)
	}

	return result
})

// 更新块元素位置缓存 - 优化版本
const updateBlockPositions = () => {
	if (!containerRef.value || !contentRef.value || blockRefs.value.length === 0) return

	const positions: BlockPositionType[] = []
	const containerRect = containerRef.value.getBoundingClientRect()

	// 使用已缓存的引用而不是查询DOM
	blockRefs.value.forEach((el, index) => {
		if (!el) return

		const rect = el.getBoundingClientRect()
		positions.push({
			index,
			top: rect.top - containerRect.top,
			bottom: rect.bottom - containerRect.top,
			center: rect.top + rect.height / 2 - containerRect.top,
			height: rect.height,
		})
	})

	// 按照位置从上到下排序
	positions.sort((a, b) => a.top - b.top)
	blockPositions.value = positions
}

// 使用二分查找确定最佳放置位置
const findBestPositionWithBinarySearch = (
	mouseY: number,
	positions: BlockPositionType[]
): number => {
	if (positions.length === 0) return 0

	// 边界情况：鼠标在第一个元素上方
	if (mouseY < positions[0].top) return 0

	// 边界情况：鼠标在最后一个元素下方
	if (mouseY > positions[positions.length - 1].bottom) return positions.length

	let start = 0
	let end = positions.length - 1

	// 二分查找
	while (start <= end) {
		const mid = Math.floor((start + end) / 2)
		const pos = positions[mid]

		// 鼠标在当前元素范围内
		if (mouseY >= pos.top && mouseY <= pos.bottom) {
			// 关键改进：根据鼠标在元素上半部分还是下半部分决定插入位置
			const middlePoint = pos.top + pos.height / 2
			return mouseY < middlePoint ? pos.index : pos.index + 1
		}

		// 调整搜索范围
		if (mouseY < pos.top) {
			end = mid - 1
		} else {
			start = mid + 1
		}
	}

	// 边界情况处理
	return start < positions.length ? start : positions.length
}

// 更新放置索引 - 优化版本
const updateDropIndexFromPosition = () => {
	if (!containerRef.value) {
		dropIndex.value = 0
		return
	}

	const containerRect = containerRef.value.getBoundingClientRect()
	const mouseY = mousePosition.value.y - containerRect.top // 使用相对于容器的位置

	// 特殊处理：容器顶部和底部的边界情况
	if (mouseY < 20) {
		dropIndex.value = 0
		return
	}

	if (mouseY > containerRect.height - 20) {
		dropIndex.value = children.value.length
		return
	}

	// 如果位置缓存不存在或为空，则重新创建
	if (!blockPositions.value.length) {
		updateBlockPositions()
		// 如果更新后仍然为空，默认放在开头
		if (!blockPositions.value.length) {
			dropIndex.value = 0
			return
		}
	}

	// 使用二分查找确定最佳位置
	dropIndex.value = findBestPositionWithBinarySearch(mouseY, blockPositions.value)
}

// 节流版的位置更新函数 - 使用更低的延迟提高响应性
const updateDropIndexThrottled = throttle(
	() => {
		updateDropIndexFromPosition()
	},
	10,
	{ leading: true, trailing: true }
)

// 重置拖放状态
const resetDropState = () => {
	isOver.value = false
	dropIndex.value = -1
	dragSourceIndex.value = -1
	dragSourceCell.value = ''
	dragItemType.value = ''
	blockPositions.value = []
}

const [, drop] = useDrop<DropBaseBlock>(() => ({
	accept: ['ADD_BASE_BLOCK', 'SORT_BASE_BLOCK'],
	hover: throttle((item, monitor) => {
		// 记录拖拽项类型
		dragItemType.value = item.type

		// 首次hover时创建位置缓存
		if (!isOver.value) {
			updateBlockPositions()
		}

		// 确保isOver状态正确设置
		if (!monitor.isOver()) {
			resetDropState()
			return
		}

		isOver.value = true

		// 记录源信息
		if (item.type === 'SORT_BASE_BLOCK') {
			dragSourceIndex.value = item.sourceIndex !== undefined ? item.sourceIndex : -1
			dragSourceCell.value = item.sourceCellKey || ''
		}

		// 获取并更新鼠标位置
		const clientOffset = monitor.getClientOffset()
		if (clientOffset) {
			mousePosition.value = {
				x: clientOffset.x,
				y: clientOffset.y,
			}

			// 节流更新位置
			updateDropIndexThrottled()
		}
	}, 16),
	drop(item) {
		if (item.type === 'ADD_BASE_BLOCK') {
			// 在指定位置插入块
			if (hasChildren.value && dropIndex.value !== -1) {
				insertBlockToCell(dropIndex.value, blockKey, item.blockData)
			} else {
				// 没有子元素或者没有有效的dropIndex，插入到末尾
				insertBlockToCell(0, blockKey, item.blockData)
			}
		}

		if (item.type === 'SORT_BASE_BLOCK') {
			// 处理块的排序
			if (dropIndex.value !== -1) {
				if (item.sourceCellKey === blockKey) {
					// 同一单元格内排序 - 检查是否拖回原位置
					const adjustedDropIndex =
						item.sourceIndex < dropIndex.value ? dropIndex.value - 1 : dropIndex.value

					// 如果不是拖回原位置，才执行移动操作
					if (item.sourceIndex !== adjustedDropIndex) {
						moveBlockInCell(blockKey, item.sourceIndex, dropIndex.value)
					}
				} else if (item.sourceCellKey) {
					// 跨单元格排序 - 移动块到新位置
					moveBlockBetweenCells(
						item.sourceCellKey,
						blockKey,
						item.sourceIndex,
						dropIndex.value,
						item.blockData.key
					)
				}
			}
		}

		// 重置状态
		resetDropState()
	},
	collect: monitor => {
		if (!monitor.isOver()) {
			resetDropState() // 使用resetDropState函数确保完全重置状态
		}
		return {}
	},
}))

// 解析块组件类型
const resolveBlockComponent = (type: BlockType) => {
	const componentMap: Record<string, Component> = {
		button: BlockButton,
		link: BlockLink,
		divider: BlockDivider,
		header: BlockHeader,
		text: BlockText,
		image: BlockImage,
		menu: BlockMenu,
		copyright: BlockCopyright,
	}

	return componentMap[type] || null
}

// 使用防抖处理滚动和调整大小事件
const handleScroll = debounce(
	() => {
		if (isOver.value) {
			updateBlockPositions() // 滚动时更新位置缓存
		}
	},
	100,
	{ leading: true, maxWait: 200 }
)

const handleResize = debounce(
	() => {
		if (isOver.value) {
			updateBlockPositions() // 调整大小时更新位置缓存
		}
	},
	100,
	{ leading: true, maxWait: 200 }
)

// 添加和移除事件监听
onMounted(() => {
	window.addEventListener('scroll', handleScroll, { passive: true })
	window.addEventListener('resize', handleResize, { passive: true })
})

onUnmounted(() => {
	window.removeEventListener('scroll', handleScroll)
	window.removeEventListener('resize', handleResize)

	// 清理节流和防抖函数
	updateDropIndexThrottled.cancel()
	handleScroll.cancel()
	handleResize.cancel()
})
</script>

<style lang="scss" scoped>
.cell-container {
	&.drop-active {
		background-color: rgba(24, 144, 255, 0.05) !important;
		border: 1px dashed #1890ff !important;
	}

	.cell-content {
		display: flex;
		flex-direction: column;
	}

	.block-item {
		position: relative;
		transition: transform 0.2s;

		&:hover {
			z-index: 1;
		}
	}

	.block-placeholder {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 0;
		overflow: hidden;
		transition: height 0.3s;
		background-color: transparent;
		border: 1px dashed transparent;
		border-radius: 4px;

		&.active {
			height: 50px;
			background-color: rgba(24, 144, 255, 0.05);
			border-color: #1890ff;
		}

		.placeholder-content {
			color: #1890ff;
			display: flex;
			flex-direction: column;
			align-items: center;
		}
	}

	.column-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100px;
		background-color: rgba(62, 158, 212, 0.15);
		border: 1px dashed rgba(62, 158, 212, 0.5);
		color: #0079d6;
		font-size: 14px;
		text-align: center;
		transition: all 0.3s;

		&.active {
			color: #1890ff;
		}
	}
}
</style>
