<template>
	<container label="Layout">
		<template #value>
			<div class="w-140px">
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
	{ label: 'Horizontal', value: 'inline-block' },
	{ label: 'Vertical', value: 'block' },
]

if (!value.value) {
	value.value = options[0].value
}
</script>
