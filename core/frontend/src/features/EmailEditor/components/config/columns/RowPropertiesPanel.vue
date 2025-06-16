<template>
	<div class="py-8px">
		<style-form v-model:value="columnsStyle"> </style-form>
	</div>
</template>

<script lang="tsx" setup>
import { useConfig } from '../../../hooks/useConfig'
import { useSetData } from '../../../hooks/useSetData'
import { useNormalForm } from '../../style/useNormalForm'

import Color from '../../style/Color.vue'

const { selectedBlockKey, columnsConfigMap } = useConfig()
const { autoSaveFn } = useSetData()

const columnsStyle = computed({
	get() {
		return columnsConfigMap.value[selectedBlockKey.value].style
	},
	set(newVal) {
		columnsConfigMap.value[selectedBlockKey.value].style = newVal
		autoSaveFn()
	},
})

const containerStyle = computed({
	get() {
		return columnsConfigMap.value[selectedBlockKey.value].containerStyle
	},
	set(newVal) {
		columnsConfigMap.value[selectedBlockKey.value].containerStyle = newVal
		autoSaveFn()
	},
})

const [StyleForm] = useNormalForm([
	{
		attrKey: 'backgroundColor',
		label: 'Background Color',
	},
	{
		attrKey: 'custom',
		render: () => {
			return (
				<Color
					label="Content Background Color"
					value={containerStyle.value.backgroundColor}
					onUpdate:value={val => {
						containerStyle.value.backgroundColor = val
					}}
				/>
			)
		},
	},
	{
		attrKey: 'padding',
		label: 'Padding',
	},
])
</script>
