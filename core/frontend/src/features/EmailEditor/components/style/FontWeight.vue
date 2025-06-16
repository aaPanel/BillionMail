<template>
	<container label="Font Weight">
		<template #value>
			<div class="w-100px">
				<n-select v-model:value="value" :options="options"></n-select>
			</div>
		</template>
	</container>
</template>

<script lang="ts" setup>
import Container from './BaseContainer.vue'

interface Props {
	label?: string
}

withDefaults(defineProps<Props>(), {
	label: '',
})

const value = defineModel<string>('value', {
	default: '',
})

const options = [
	{ label: 'Normal', value: 'normal' },
	{ label: 'Bold', value: 'bold' },
]

if (!value.value) {
	value.value = options[0].value
}
</script>
