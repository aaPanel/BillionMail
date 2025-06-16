<template>
	<container label="Width">
		<template #value>
			<span class="mr-12px">Auto On</span>
			<n-switch v-model:value="auto" @update:value="onUpdateAuto"></n-switch>
		</template>
		<template #options>
			<div class="flex items-center h-32px">
				<n-slider v-model:value="width" :step="1" :disabled="auto" @update:value="onUpdateWidth" />
				<span v-if="!auto" class="ml-8px whitespace-nowrap">{{ width }}%</span>
			</div>
		</template>
	</container>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue'
import Container from './BaseContainer.vue'

const value = defineModel<string>('value', {
	default: '',
})

const auto = ref(true)

const width = ref(100)

watch(
	() => value.value,
	val => {
		if (val === 'auto') {
			auto.value = true
		} else {
			auto.value = false
			width.value = Number.parseFloat(val)
		}
		if (!auto.value) {
			width.value = Number.parseFloat(val)
		}
	},
	{
		immediate: true,
	}
)

const onUpdateAuto = (val: boolean) => {
	width.value = 100
	if (val) {
		value.value = '100%'
	} else {
		value.value = `${width.value}%`
	}
}

const onUpdateWidth = (val: number) => {
	value.value = `${val}%`
}
</script>
