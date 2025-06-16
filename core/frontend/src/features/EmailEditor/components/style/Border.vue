<template>
	<container :label="label || 'Border'">
		<template #value>
			<span class="mr-12px">More Options</span>
			<n-switch v-model:value="config.more" @update:value="onUpdateMore"></n-switch>
		</template>
		<template #options>
			<more-options :show="config.more">
				<template #all>
					<base-border v-model:value="config.all" />
				</template>
				<template #top>
					<base-border v-model:value="config.top" />
				</template>
				<template #left>
					<base-border v-model:value="config.left" />
				</template>
				<template #right>
					<base-border v-model:value="config.right" />
				</template>
				<template #bottom>
					<base-border v-model:value="config.bottom" />
				</template>
			</more-options>
		</template>
	</container>
</template>

<script lang="ts" setup>
import { BorderStyle, MoreStyle } from '../../types/base'

import Container from './BaseContainer.vue'
import MoreOptions from './MoreOptions.vue'
import BaseBorder from './BaseBorder.vue'

interface Props {
	label?: string
}

withDefaults(defineProps<Props>(), {
	label: '',
})

const config = defineModel<MoreStyle<BorderStyle>>('value', {
	default: () => ({
		more: false,
		all: {
			width: '0',
			style: 'solid',
			color: '',
		},
		top: {
			width: '0',
			style: 'solid',
			color: '',
		},
		left: {
			width: '0',
			style: 'solid',
			color: '',
		},
		right: {
			width: '0',
			style: 'solid',
			color: '',
		},
		bottom: {
			width: '0',
			style: 'solid',
			color: '',
		},
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
