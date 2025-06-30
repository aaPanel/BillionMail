<template>
	<div class="flex">
		<div class="w-0 flex-1 flex items-center">
			<div class="min-w-0">
				<n-ellipsis class="max-w-full" :tooltip="{ contentStyle: { maxWidth: '300px' } }">
					{{ value }}
				</n-ellipsis>
			</div>
			<div class="ml-6px text-15px cursor-pointer" title="Copy" @click="onCopy">
				<i class="i-mdi-content-copy"></i>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useCopy } from '@/hooks/useCopy'

interface Props {
	value: string | number
}

const props = withDefaults(defineProps<Props>(), {
	value: '',
})

const { copyText } = useCopy()

const onCopy = async () => {
	copyText(`${props.value}`)
}
</script>
