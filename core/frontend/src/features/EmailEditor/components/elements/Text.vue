<template>
	<block-section :data="data" :cell-key="cellKey" :block-type="data.type" :block-index="blockIndex"
		@select="handleSelect" @unSelect="handleUnSelect">
		<div :style="textStyle">
			<text-editor ref="editorRef" v-model:value="info.content" type="text" />
		</div>
	</block-section>
</template>

<script lang="ts" setup>
import { useConfig } from '../../hooks/useConfig'
import { useStyle } from '../../hooks/useStyle'
import { BlockBaseType } from '../../types/base'

import BlockSection from '../layout/Section.vue'
import TextEditor from '../shared/Editor.vue'

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

const editorRef = useTemplateRef('editorRef')

const info = computed({
	get() {
		return blockConfigMap.value[data.key].attr
	},
	set(value) {
		blockConfigMap.value[data.key].attr = value
	},
})

const textStyle = computed(() => {
	const { style } = blockConfigMap.value[data.key]
	return {
		...configToStyle(style),
		width: '100%',
	}
})

const handleSelect = () => {
	editorRef.value?.focus()
}

const handleUnSelect = () => {
	editorRef.value?.blur()
}
</script>
