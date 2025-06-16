<template>
	<Container label="Link">
		<template #options>
			<div>
				<n-input v-model:value="value" :disabled="isUnsubscribe"></n-input>
			</div>
			<div class="mt-12px">
				<n-switch v-model:value="target" :checked-value="'_blank'" :unchecked-value="'_self'">
				</n-switch>
				<span class="ml-8px">Open link in new tab</span>
			</div>
			<div class="mt-12px">
				<n-switch v-model:value="isUnsubscribe" @update:value="handleUnsubscribe"> </n-switch>
				<span class="ml-8px">Unsubscribe</span>
			</div>
		</template>
	</Container>
</template>

<script lang="ts" setup>
import Container from './BaseContainer.vue'

const value = defineModel<string>('value')

const target = defineModel<string>('target')

// 退订
const isUnsubscribe = ref(false)

const handleUnsubscribe = (val: boolean) => {
	if (val) {
		value.value = '{{ UnsubscribeURL . }}'
	} else {
		value.value = ''
	}
}
</script>
