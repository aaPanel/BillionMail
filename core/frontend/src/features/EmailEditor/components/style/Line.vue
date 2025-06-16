<template>
	<container label="Line">
		<template #value>
			<base-border v-model:value="value"></base-border>
		</template>
	</container>
</template>

<script lang="ts" setup>
import { BorderStyle } from '../../types/base'
import Container from './BaseContainer.vue'
import BaseBorder from './BaseBorder.vue'

const value = defineModel<BorderStyle>('value', {
	default: () => ({
		width: '0',
		style: 'solid',
		color: '',
	}),
})
</script>
