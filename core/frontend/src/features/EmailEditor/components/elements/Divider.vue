<template>
	<block-section :data="data" :cell-key="cellKey" :block-type="data.type" :block-index="blockIndex">
		<hr :style="style"></hr>
	</block-section>
</template>

<script lang="ts" setup>
import { useConfig } from '../../hooks/useConfig'
import { useStyle } from '../../hooks/useStyle'
import { BlockBaseType } from '../../types/base'

import BlockSection from '../layout/Section.vue'

const { data } = defineProps({
	data: {
		type: Object as PropType<BlockBaseType>,
		required: true,
	},
	cellKey: {
		type: String,
		required: true,
	},
	blockIndex: {
		type: Number,
		required: true,
	},
})

const { blockConfigMap } = useConfig()
const { configToStyle } = useStyle()

const style = computed(() => {
	const { style } = blockConfigMap.value[data.key]
	return {
		...configToStyle(style),
	}
})
</script>
