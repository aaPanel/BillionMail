<template>
	<container label="Rounded Border">
		<template #value>
			<span class="mr-12px">More Options</span>
			<n-switch v-model:value="config.more" @update:value="onUpdateMore"></n-switch>
		</template>
		<template #options>
			<more-options
				:show="config.more"
				top-label="Top Left"
				left-label="Bottom Left"
				right-label="Top Right"
				bottom-label="Bottom Right">
				<template #all>
					<base-number v-model:value="config.all" />
				</template>
				<template #top>
					<base-number v-model:value="config.top" />
				</template>
				<template #left>
					<base-number v-model:value="config.left" />
				</template>
				<template #right>
					<base-number v-model:value="config.right" />
				</template>
				<template #bottom>
					<base-number v-model:value="config.bottom" />
				</template>
			</more-options>
		</template>
	</container>
</template>

<script lang="ts" setup>
import { MoreStyle } from '../../types/base'

import Container from './BaseContainer.vue'
import BaseNumber from './BaseNumber.vue'
import MoreOptions from './MoreOptions.vue'

const config = defineModel<MoreStyle<string>>('value', {
	default: () => ({
		more: false,
		all: '',
		top: '',
		left: '',
		right: '',
		bottom: '',
	}),
})

const onUpdateMore = (val: boolean) => {
	if (val) {
		config.value.top = config.value.all
		config.value.right = config.value.all
		config.value.bottom = config.value.all
		config.value.left = config.value.all
	} else {
		config.value.all = config.value.top
	}
}
</script>
