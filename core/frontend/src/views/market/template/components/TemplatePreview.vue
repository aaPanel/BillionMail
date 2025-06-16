<template>
	<modal :title="$t('market.template.preview.title')" width="820" :footer="false">
		<bt-preview :height="580" :value="content"></bt-preview>
	</modal>
</template>

<script lang="ts" setup>
import { useModal } from '@/hooks/modal/useModal'

const content = ref('')

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ html: string }>()
			content.value = state.html || ''
		}
	},
})
</script>
