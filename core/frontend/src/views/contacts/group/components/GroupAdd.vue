<template>
	<modal :title="t('contacts.group.add.title')" width="500">
		<div class="pt-12px">
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-form-item :label="t('contacts.group.form.name')" path="name">
					<n-input v-model:value="form.name"></n-input>
				</n-form-item>
				<n-form-item label="Status">
					<n-select v-model:value="form.status" :options="statusOptions"></n-select>
				</n-form-item>
				<n-form-item :label="t('contacts.group.form.file')" path="file_data">
					<div class="flex-1">
						<bt-file-upload :is-upload="false" :accept="['csv']" @change="handleChangeFile">
						</bt-file-upload>
						<div class="mt-8px text-desc">
							<span class="mr-4px">{{ t('contacts.group.fileUpload.hint') }}</span>
							<n-button text type="primary" @click="handleDownloadTemplate">
								{{ $t('common.actions.download') }}
							</n-button>
						</div>
					</div>
				</n-form-item>
				<n-form-item label="Example raw CSV">
					<pre class="csv-example">{{ example }}</pre>
				</n-form-item>
			</bt-form>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { UploadFileInfo } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { downloadFile } from '@/api/modules/public'
import { createGroup } from '@/api/modules/contacts/group'

const { t } = useI18n()

const formRef = useTemplateRef('formRef')

const form = reactive({
	create_type: 2,
	name: '',
	description: '',
	status: 1,
	file_data: '',
	file_type: '',
})

const rules = {
	name: {
		required: true,
		message: t('contacts.group.form.nameRequired'),
		trigger: 'blur',
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

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			form.name = ''
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()
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
