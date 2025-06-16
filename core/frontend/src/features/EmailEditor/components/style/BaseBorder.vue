<!--
	@description 
		- 基础边框设置，为Border服务（可以单独设置四条边框）  
		- 要求数据格式为字符串
-->
<template>
	<div class="flex flex-col gap-8px">
		<number-base v-model:value="config.width" placeholder=""> </number-base>
		<div class="w-92px">
			<n-select v-model:value="config.style" :options="options" placeholder=""> </n-select>
		</div>
		<div class="w-32px">
			<n-color-picker v-model:value="config.color" :modes="['hex']">
				<template #label><span></span></template>
			</n-color-picker>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { BorderStyle } from '../../types/base'
import NumberBase from './BaseNumber.vue'

const config = defineModel<BorderStyle>('value', {
	default: () => ({
		width: '0',
		style: 'solid',
		color: '',
	}),
})

const options = [
	{ label: 'Solid', value: 'solid' },
	{ label: 'Dotted', value: 'dotted' },
	{ label: 'Dashed', value: 'dashed' },
]
</script>
