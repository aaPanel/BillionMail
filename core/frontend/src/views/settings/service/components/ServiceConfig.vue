<template>
	<modal :title="title" :width="800">
		<div class="mb-16px">
			<div class="text-desc">
				{{ t('settings.service.tips') }}
			</div>
		</div>
		<div class="h-560px">
			<bt-editor v-model:value="content" @save="saveConfig"></bt-editor>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { isObject } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { getServiceConfig, saveServiceConfig } from '@/api/modules/settings'
import type { DockerService } from '../types/common'

const { t } = useI18n()

const name = ref('')

const content = ref('')

const title = computed(() => {
	return t('settings.service.editTitle', { name: name.value })
})

const getConfig = async () => {
	const res = await getServiceConfig({ service_type: name.value })
	if (isObject<{ content: string }>(res)) {
		content.value = res.content
	}
}

const saveConfig = async () => {
	await saveServiceConfig({
		service_type: name.value,
		content: content.value,
	})
}

const [Modal, modalApi] = useModal({
	onChangeState: async isOpen => {
		if (isOpen) {
			const { row } = modalApi.getState<{ row: DockerService }>()
			name.value = row.Names[0].trim()
			getConfig()
		} else {
			name.value = ''
			content.value = ''
		}
	},
	onConfirm: async () => {
		await saveConfig()
	},
})
</script>

<style lang="scss" scoped></style>
