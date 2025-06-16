<template>
	<container label="Width">
		<template #value>
			<!-- <base-number v-model:value="value"></base-number> -->
			<div class="w-180px">
				<n-slider v-model:value="width" :step="1" @update:value="onUpdateWidth" />
			</div>
		</template>
	</container>
</template>

<script lang="ts" setup>
import { ref, watch } from "vue"
import Container from './BaseContainer.vue'

const value = defineModel<string>('value', {
	default: '',
})

const width = ref(100)

watch(
	() => value.value,
	val => {
		width.value = Number.parseFloat(val)
	},
	{
		immediate: true,
	}
)

const onUpdateWidth = (val: number) => {
	value.value = `${val}%`
}
</script>
