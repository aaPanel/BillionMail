<template>
	<container label="Letter Spacing">
		<template #value>
			<base-number v-model:value="value"></base-number>
		</template>
	</container>
</template>

<script lang="ts" setup>
import Container from './BaseContainer.vue'
import BaseNumber from './BaseNumber.vue'

const value = defineModel<string>('value', {
	default: '0px',
})
</script>
