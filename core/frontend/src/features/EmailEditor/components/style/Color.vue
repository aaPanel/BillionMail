<template>
	<container :label="label">
		<template #value>
			<div class="w-32px">
				<n-color-picker v-model:value="color" :modes="['hex']">
					<template #label><span></span></template>
				</n-color-picker>
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

const color = defineModel<string>('value', {
	default: '',
})
</script>
