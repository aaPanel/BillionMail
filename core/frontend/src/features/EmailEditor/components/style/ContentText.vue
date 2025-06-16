<template>
	<Container label="Content">
		<template #options>
			<div>
				<n-input v-model:value="value"></n-input>
			</div>
		</template>
	</Container>
</template>

<script lang="ts" setup>
import Container from './BaseContainer.vue'

const value = defineModel<string>('value')
</script>
