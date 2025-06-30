<template>
	<modal :title="t('market.task.createGroup.title')" width="500">
		<div class="pt-12px">
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-form-item :label="t('market.task.createGroup.groupName')" path="name">
					<n-input
						v-model:value="form.name"
						:placeholder="t('market.task.createGroup.groupNamePlaceholder')">
					</n-input>
				</n-form-item>
				<n-form-item label="Status">
					<n-select v-model:value="form.status" :options="statusOptions"></n-select>
				</n-form-item>
				<n-form-item :label="t('market.task.createGroup.email')" path="file_data">
					<div class="flex-1">
						<bt-file-upload :is-upload="false" :accept="['csv']" @change="handleChangeFile">
						</bt-file-upload>
						<div class="mt-8px text-12px text-desc">
							{{ $t('contacts.subscribers.import.fileTypeHint') }}
							<n-button text type="primary" @click="handleDownloadTemplate">
								{{ $t('common.actions.download') }}
							</n-button>
						</div>
					</div>
				</n-form-item>
				<n-form-item :label="$t('contacts.subscribers.import.exampleCsvLabel')">
					<pre class="csv-example">{{ example }}</pre>
				</n-form-item>
			</bt-form>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { isEmpty } from 'lodash-es'
import { FormRules, UploadFileInfo } from 'naive-ui'
import { Message } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { downloadFile } from '@/api/modules/public'
import { createGroup } from '@/api/modules/contacts/group'

const { t } = useI18n()

const formRef = useTemplateRef('formRef')

const form = reactive({
	create_type: 2,
	name: '',
	file_data: '',
	file_type: '',
	description: '',
	status: 1,
})

const rules: FormRules = {
	name: {
		required: true,
		message: t('market.task.createGroup.validation.groupNameRequired'),
		trigger: ['blur', 'input'],
	},
}

const statusOptions = [
	{ label: 'Confirmed', value: 1 },
	{ label: 'Unconfirmed', value: 0 },
]

const example = `email,attributes
example1@example.com,"{""age"": ""13"", ""Gender"": ""XXX""}"
example2@example.com,"{""age"": ""45"", ""city"": ""XXX""}"
`

const handleChangeFile = (file: UploadFileInfo) => {
	const reader = new FileReader()
	reader.onload = e => {
		form.file_type = file.name.substring(file.name.lastIndexOf('.') + 1)
		form.file_data = e.target?.result as string
	}
	if (file.file) {
		reader.readAsText(file.file)
	}
}

const handleDownloadTemplate = async () => {
	await downloadFile({ file_path: './template/example_recipients.csv' })
}

const validateForm = async () => {
	await formRef.value?.validate()
	if (isEmpty(form.file_data)) {
		Message.error(t('contacts.subscribers.import.validation.fileRequired'))
		return Promise.reject()
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			form.name = ''
			form.file_data = ''
			form.file_type = ''
		}
	},
	onConfirm: async () => {
		await validateForm()
		await createGroup(toRaw(form))
		const state = modalApi.getState<{ refresh: () => void }>()
		state?.refresh()
	},
})
</script>

<style lang="scss" scoped>
.csv-example {
	width: 100%;
	margin: 0;
	padding: 16px 24px;
	background: none;
	border: 1px solid var(--color-border-1);
	border-radius: 4px;
	overflow-x: auto;
	white-space: pre;
	word-wrap: normal;
}
</style>
