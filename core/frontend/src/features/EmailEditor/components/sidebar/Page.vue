<template>
	<div class="p-20px">
		<StyleForm v-model:value="pageStyle" />
	</div>
</template>

<script setup lang="ts">
import { useConfig } from '../../hooks/useConfig'
import { useNormalForm } from '../style/useNormalForm'

const { pageConfig } = useConfig()

const pageStyle = computed({
	get() {
		return pageConfig.value.style
	},
	set(newVal) {
		pageConfig.value.style = newVal
	},
})

const [StyleForm] = useNormalForm([
	{
		attrKey: 'backgroundColor',
		label: 'Background Color',
	},
	{
		attrKey: 'widthNumber',
		label: 'Width',
		key: 'width',
	},
])
</script>
