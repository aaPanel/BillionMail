<template>
	<block-section :data="data" :cell-key="cellKey" :block-type="data.type" :block-index="blockIndex">
		<div>
			<None v-if="info.links?.length === 0">Menu</None>
			<a v-for="item of info.links" :key="item.id" :style="style">{{ item.label }}</a>
		</div>
	</block-section>
</template>

<script lang="ts" setup>
import { useConfig } from '../../hooks/useConfig'
import { useStyle } from '../../hooks/useStyle'
import { BlockBaseType } from '../../types/base'

import BlockSection from '../layout/Section.vue'
import None from '../shared/None.vue'

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

const info = computed(() => {
	return blockConfigMap.value[data.key].attr
})
</script>
