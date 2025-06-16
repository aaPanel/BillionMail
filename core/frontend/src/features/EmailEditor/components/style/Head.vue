<template>
	<container label="Heading Type">
		<template #value>
			<n-radio-group v-model:value="value">
				<n-radio-button value="22px">H1</n-radio-button>
				<n-radio-button value="20px">H2</n-radio-button>
				<n-radio-button value="18px">H3</n-radio-button>
				<n-radio-button value="16px">H4</n-radio-button>
			</n-radio-group>
		</template>
	</container>
</template>

<script lang="ts" setup>
import Container from './BaseContainer.vue'

const value = defineModel<string>('value')
</script>

<style lang="scss" scoped></style>
