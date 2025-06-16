<template>
	<container label="Line Height">
		<template #value>
			<base-number v-model:value="value" unit="%" :step="10"></base-number>
		</template>
	</container>
</template>

<script lang="ts" setup>
import Container from './BaseContainer.vue'
import BaseNumber from './BaseNumber.vue'

const value = defineModel<string>('value', {
	default: '120%',
})
</script>
