<template>
	<div
		:ref="preview"
		class="block-base"
		:class="[`block-${blockType}`, { 'is-selected': isSelected }, { 'is-dragging': isDragging }]"
		:data-block-id="data.key"
		:style="sectionStyle"
		@click.stop="handleSelect">
		<div class="block-section">
			<!-- 选中工具栏 -->
			<BlockToolbar v-show="isSelected && !isDragging" :cell-key="cellKey" :data="data" />
			<!-- 悬浮提示 -->
			<!-- <div v-show="isHovered && !isSelected" class="block-type-hint">
			{{ getDisplayName() }}
		</div> -->
			<!-- 列组件拖拽排序按钮 -->
			<div
				v-if="data.type === 'columns'"
				:ref="columnsDrag"
				class="drag-handle columns-drag-handle">
				<i class="drag-icon">≡</i>
			</div>
			<!-- 基础组件拖拽排序按钮 -->
			<div v-if="data.type !== 'columns'" :ref="drag" class="drag-handle">
				<i class="drag-icon">≡</i>
			</div>
		</div>
		<div class="block-content">
			<slot></slot>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useDrag } from 'vue3-dnd'
import { cloneDeep } from 'lodash-es'
import { useStyle } from '../../hooks/useStyle'
import { useConfig } from '../../hooks/useConfig'
import { BlockBaseType, BlockType } from '../../types/base'

import BlockToolbar from './Toolbar.vue'

const { data, cellKey, blockIndex, blockType } = defineProps({
	data: {
		type: Object as PropType<BlockBaseType>,
		required: true,
	},
	cellKey: {
		type: String,
		default: '',
	},
	blockType: {
		type: String as PropType<BlockType>,
		required: true,
	},
	blockIndex: {
		type: Number,
		default: -1,
	},
})

const emit = defineEmits<{
	select: []
	unSelect: []
}>()

const { selectedBlockKey, columnsConfigMap, blockConfigMap, selectBlock } = useConfig()
const { configToStyle } = useStyle()

// 判断当前组件是否被选中
const isSelected = computed(() => {
	return selectedBlockKey.value === data.key
})

// 处理组件选中事件
const handleSelect = () => {
	selectBlock(data.key, data.type)
	emit('select')
}

const sectionStyle = computed(() => {
	let style = {}
	if (data.type === 'columns') {
		style = columnsConfigMap.value[data.key].containerStyle
	} else {
		style = blockConfigMap.value[data.key].containerStyle
	}
	return configToStyle(style)
})

// 获取显示名称
// const getDisplayName = () => {
// 	return props.data.name
// }

// 使用vue3-dnd的useDrag hook
const [baseCollect, drag, preview] = useDrag(() => ({
	type: 'SORT_BASE_BLOCK',
	item: () => ({
		type: 'SORT_BASE_BLOCK',
		sourceIndex: blockIndex,
		sourceCellKey: cellKey,
		blockData: cloneDeep(data),
	}),
	collect: monitor => ({
		isDragging: monitor.isDragging(),
	}),
}))

const [columnsCollect, columnsDrag] = useDrag(() => ({
	type: 'SORT_COLUMNS_BLOCK',
	item: () => ({
		type: 'SORT_COLUMNS_BLOCK',
		sourceIndex: blockIndex,
		blockData: cloneDeep(data),
	}),
	collect: monitor => ({
		isDragging: monitor.isDragging(),
	}),
}))

const isDragging = computed(() => {
	return baseCollect.value.isDragging || columnsCollect.value.isDragging
})

watch(
	() => isDragging.value,
	newVal => {
		if (newVal === true) {
			emit('unSelect')
		}
	}
)

watch(
	() => isSelected.value,
	(newVal, oldVal) => {
		if (newVal === false && oldVal === true) {
			emit('unSelect')
		}
	}
)
</script>

<style lang="scss" scoped>
.block-base {
	position: relative;

	.block-section {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		width: 100%;
		user-select: none;

		&::after {
			content: '';
			display: block;
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;
			opacity: 0;
			background-color: transparent;
			z-index: 100;
			pointer-events: none;
		}

		&:hover {
			&::after {
				opacity: 1;
				border: 2px solid #4eaef8;
			}

			.drag-handle {
				opacity: 1;
			}
		}
	}

	&.is-selected {
		> .block-section {
			&::after {
				opacity: 1;
				border: 2px solid #0871c2;
			}

			.drag-handle {
				opacity: 1;
			}
		}
	}

	&.is-dragging {
		opacity: 0.4;
	}

	.block-content {
		margin: 0 auto;
	}

	.block-type-hint {
		position: absolute;
		bottom: -24px;
		left: 0;
		background: rgba(0, 0, 0, 0.6);
		color: white;
		padding: 2px 8px;
		border-radius: 3px;
		font-size: 12px;
		opacity: 0.8;
		pointer-events: none;
		z-index: 102;
	}

	// 拖拽手柄样式
	.drag-handle {
		position: absolute;
		right: -12px;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border-radius: 3px;
		background-color: #fff;
		border: 1px solid #e8e8e8;
		cursor: grab;
		z-index: 101;
		color: #909399;
		opacity: 0;
		transition: opacity 0.2s;

		&.columns-drag-handle {
			right: 0;
		}

		&:hover {
			color: #1890ff;
		}

		.drag-icon {
			font-style: normal;
			font-size: 16px;
		}
	}
}
</style>
