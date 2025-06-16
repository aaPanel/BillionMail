<template>
	<div class="block-toolbar">
		<button class="tool-btn" @click.stop="duplicate">
			<i class="i-ri:file-copy-line text-16px text-#fff"></i>
		</button>
		<button class="tool-btn" @click.stop="remove">
			<i class="i-mdi:delete-outline text-16px text-#fff"></i>
		</button>
	</div>
</template>

<script lang="ts" setup>
import { useConfig } from '../../hooks/useConfig'
import { BlockBaseType } from '../../types/base'

const { cellKey, data } = defineProps({
	cellKey: {
		type: String,
		required: true,
	},
	data: {
		type: Object as PropType<BlockBaseType>,
		required: true,
	},
})

const { duplicateBlock, duplicateColumn, removeBlock, removeColumn } = useConfig()

const duplicate = () => {
	if (data.type !== 'columns') {
		duplicateBlock(cellKey, data.key)
	} else {
		duplicateColumn(data.key)
	}
}

const remove = () => {
	if (data.type !== 'columns') {
		removeBlock(cellKey, data.key)
	} else {
		removeColumn(data.key)
	}
}
</script>

<style lang="scss" scoped>
.block-toolbar {
	position: absolute;
	bottom: -36px;
	right: 0;
	display: flex;
	background-color: #1890ff;
	border-radius: 4px 4px 0 0;
	z-index: 101;

	.tool-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		padding: 0;
		border: none;
		background: transparent;
		color: white;
		cursor: pointer;
		transition: background-color 0.2s;

		svg {
			width: 16px;
			height: 16px;
		}

		&:hover {
			background-color: rgba(0, 0, 0, 0.1);
		}
	}
}
</style>
