<template>
	<block-section :data="data" :cell-key="cellKey" :block-type="data.type" :block-index="blockIndex">
		<div :style="style">
			<img v-if="info.src" :src="info.src" :alt="info.alt" />
			<None v-else>Image</None>
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

const info = computed({
	get() {
		return blockConfigMap.value[data.key].attr
	},
	set(value) {
		blockConfigMap.value[data.key].attr = value
	},
})
</script>

<style lang="scss" scoped>
img {
	width: 100%;
}
</style>
