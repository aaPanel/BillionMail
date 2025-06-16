<template>
	<div class="bg-gray-900 text-gray-100 rounded-lg p-4 pr-48px font-mono text-sm relative">
		<span class="text-gray-400 mr-2">$</span>{{ value }}
		<div class="absolute top-0px right-16px h-full flex items-center">
			<n-button text class="text-desc" @click="copyText(value)">
				<i class="i-mdi-content-copy text-18px"></i>
			</n-button>
		</div>
	</div>
	<p class="mt-2 text-12px text-desc">{{ $t('settings.common.system.copyCommand') }}</p>
</template>

<script lang="ts" setup>
import { useCopy } from '@/hooks/useCopy'

defineProps({
	value: {
		type: String,
		default: '',
	},
})

const { copyText } = useCopy()
</script>

<style lang="scss" scoped></style>
