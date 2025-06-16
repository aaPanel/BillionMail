<template>
	<div :ref="drag" class="block-item">
		<div class="icon">
			<component :is="iconMap[type]"></component>
		</div>
		<div class="text">{{ data.name }}</div>
	</div>
</template>

<script lang="tsx" setup>
import { VNodeChild } from 'vue'
import { useDrag } from 'vue3-dnd'
import { cloneDeep } from 'lodash-es'
import { BlockType, MenuBlock } from '../../types/base'

const { type, data } = defineProps({
	type: {
		type: String as PropType<BlockType>,
		required: true,
	},
	data: {
		type: Object as PropType<MenuBlock>,
		required: true,
	},
})

const [, drag] = useDrag(() => {
	return {
		type: type === 'columns' ? 'ADD_COLUMNS_BLOCK' : 'ADD_BASE_BLOCK',
		item: {
			type: type === 'columns' ? 'ADD_COLUMNS_BLOCK' : 'ADD_BASE_BLOCK',
			blockData: cloneDeep(data),
		},
	}
})

const iconMap: Record<BlockType, () => VNodeChild> = {
	columns: () => <i class="i-mdi:view-column-outline"></i>,
	button: () => <i class="i-mdi:button-outline"></i>,
	link: () => <i class="i-mdi:link"></i>,
	divider: () => <i class="i-mdi:horizontal-line"></i>,
	header: () => <i class="i-ri:heading"></i>,
	text: () => <i class="i-mdi:format-text"></i>,
	image: () => <i class="i-mdi:file-image"></i>,
	menu: () => <i class="i-mdi:menu"></i>,
	placeholder: () => <i></i>,
	cell: () => <i></i>,
	html: () => <i></i>,
}
</script>

<style lang="scss" scoped>
.block-item {
	position: relative;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 8px;
	height: 74px;
	border: 1px solid #d4d4d4;
	background-color: #fff;
	border-radius: 2px;
	text-align: center;
	font-weight: 500;
	cursor: grab;
	transition: box-shadow 0.1s;

	.icon {
		width: 26px;
		height: 26px;
		font-size: 26px;
	}

	&:hover {
		box-shadow: rgba(0, 0, 0, 0.17) -2px 7px 20px;
		border: 1px solid #eeeeee;
	}
}
</style>
