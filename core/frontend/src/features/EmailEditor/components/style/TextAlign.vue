<template>
	<container :label="label || 'Text Align'">
		<template #value>
			<n-radio-group v-model:value="value">
				<n-radio-button value="left">Left</n-radio-button>
				<n-radio-button value="center">Center</n-radio-button>
				<n-radio-button value="right">Right</n-radio-button>
			</n-radio-group>
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

const value = defineModel<string>('value')
</script>
