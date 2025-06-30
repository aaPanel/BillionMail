<template>
	<modal title="Test API" width="440">
		<bt-form ref="formRef" :model="form" :rules="rules" class="pt-12px">
			<n-form-item :label="$t('market.task.edit.testEmail')" path="recipient">
				<div class="flex-1 mr-10px">
					<n-input
						v-model:value="form.recipient"
						:placeholder="$t('market.task.edit.testEmailPlaceholder')">
					</n-input>
				</div>
			</n-form-item>
		</bt-form>
	</modal>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { testApi } from '@/api/modules/api'
import type { Api } from '../types/base'

const { t } = useI18n()

const formRef = useTemplateRef('formRef')

const form = reactive({
	key: '',
	recipient: '',
})

const rules: FormRules = {
	recipient: {
		required: true,
		message: t('market.task.edit.testEmailPlaceholder'),
	},
}

const resetForm = () => {
	form.key = ''
	form.recipient = ''
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ row: Api }>()
			const { row } = state
			if (row) {
				form.key = row.api_key
			}
		} else {
			resetForm()
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()
		await testApi(form.key, {
			recipient: form.recipient,
		})
	},
})
</script>

<style lang="scss" scoped></style>
