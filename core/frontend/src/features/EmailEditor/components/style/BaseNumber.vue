<template>
	<n-input-group class="w-130px">
		<n-input-number
			v-model:value="number"
			:min="0"
			:step="step"
			placeholder=""
			@update:value="onUpdateNumber">
		</n-input-number>
		<n-input-group-label class="w-40px text-center">{{ unit }}</n-input-group-label>
	</n-input-group>
</template>

<script lang="ts" setup>
const { unit } = defineProps({
	unit: {
		type: String,
		default: 'px',
	},
	step: {
		type: Number,
		default: 1,
	},
})

const value = defineModel<string>('value', {
	default: '',
})

const number = ref(0)

watch(
	() => value.value,
	val => {
		number.value = val ? Number.parseFloat(val) : 0
	},
	{
		immediate: true,
	}
)

const onUpdateNumber = (val: number | null) => {
	value.value = `${val || 0}${unit}`
}
</script>
