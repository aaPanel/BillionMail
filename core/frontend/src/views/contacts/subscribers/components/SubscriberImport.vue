<template>
	<modal :title="$t('contacts.subscribers.import.title')" width="600">
		<div class="pt-12px">
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-form-item :label="$t('contacts.subscribers.import.overwriteLabel')">
					<div class="flex-1">
						<div class="mb-8px text-desc">
							{{ $t('contacts.subscribers.import.overwriteDescription') }}
						</div>
						<div class="w-260px">
							<n-select v-model:value="form.overwrite" :options="overwriteOptions"></n-select>
						</div>
					</div>
				</n-form-item>
				<n-grid :cols="2" :x-gap="24">
					<n-form-item-gi :span="1" :label="$t('contacts.subscribers.import.subscriptionStatus')">
						<n-select v-model:value="form.default_active" :options="modeOptions"></n-select>
					</n-form-item-gi>
					<n-form-item-gi :span="1" :label="$t('contacts.subscribers.import.status')">
						<n-select v-model:value="form.status" :options="statusOptions"></n-select>
					</n-form-item-gi>
				</n-grid>
				<n-form-item :label="$t('contacts.subscribers.import.group')" path="group_id">
					<group-select v-model:value="form.group_ids"></group-select>
				</n-form-item>
				<n-form-item :label="$t('contacts.subscribers.import.data')" path="contacts">
					<div class="flex-1">
						<n-radio-group v-model:value="form.import_type">
							<n-radio-button :value="1">
								{{ $t('contacts.subscribers.import.uploadFile') }}
							</n-radio-button>
							<n-radio-button :value="2">
								{{ $t('contacts.subscribers.import.pasteData') }}
							</n-radio-button>
						</n-radio-group>
						<div class="mt-8px">
							<template v-if="form.import_type === 1">
								<bt-file-upload :is-upload="false" :accept="['csv']" @change="handleChangeFile">
								</bt-file-upload>
								<div class="mt-8px text-12px text-desc">
									{{ $t('contacts.subscribers.import.fileTypeHint') }}
									<n-button text type="primary" @click="handleDownloadTemplate">
										{{ $t('common.actions.download') }}
									</n-button>
								</div>
							</template>
							<template v-if="form.import_type === 2">
								<n-input v-model:value="form.contacts" type="textarea" :rows="8"> </n-input>
								<div class="mt-8px text-12px text-desc">
									<p>{{ $t('contacts.subscribers.import.pastePlaceholder.0') }}</p>
									<!-- <p>{{ $t('contacts.subscribers.import.pastePlaceholder.1') }}</p> -->
									<!-- <p>{{ $t('contacts.subscribers.import.pastePlaceholder.2') }}</p> -->
								</div>
							</template>
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
import { FormRules, UploadFileInfo } from 'naive-ui'
import { isEmpty } from 'lodash-es'
import { Message } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { downloadFile } from '@/api/modules/public'
import { importSubscribers } from '@/api/modules/contacts/subscribers'

import GroupSelect from './GroupMultipleSelect.vue'

const { t } = useI18n()

const formRef = useTemplateRef('formRef')

const dataType = ref('paste')

const form = reactive({
	group_ids: [] as number[],
	import_type: 1,
	contacts: '',
	file_data: '',
	file_type: '',
	overwrite: 0,
	default_active: 1,
	status: 1,
})

const rules: FormRules = {
	group_ids: {
		trigger: 'change',
		validator: () => {
			if (form.group_ids.length === 0) {
				return new Error(t('contacts.subscribers.import.validation.groupRequired'))
			}
			return true
		},
	},
	contacts: {
		validator: () => {
			if (form.import_type === 2 && isEmpty(form.contacts)) {
				return new Error(t('contacts.subscribers.import.validation.dataRequired'))
			}
			return true
		},
	},
}

const overwriteOptions = computed(() => [
	{ label: t('contacts.subscribers.import.overwriteOptions.no'), value: 0 },
	{ label: t('contacts.subscribers.import.overwriteOptions.yes'), value: 1 },
])

const modeOptions = computed(() => [
	{ label: t('contacts.subscribers.import.subscriptionOptions.subscribe'), value: 1 },
	{ label: t('contacts.subscribers.import.subscriptionOptions.unsubscribe'), value: 0 },
])

const statusOptions = computed(() => [
	{ label: t('contacts.subscribers.import.statusOptions.confirmed'), value: 1 },
	{ label: t('contacts.subscribers.import.statusOptions.unconfirmed'), value: 0 },
])

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
	if (dataType.value === 'upload' && isEmpty(form.file_data)) {
		Message.error(t('contacts.subscribers.import.validation.fileRequired'))
		return Promise.reject()
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ group_id: number }>()
			if (state?.group_id) {
				form.group_ids = [state.group_id]
			} else {
				form.group_ids = []
			}
		} else {
			form.import_type = 1
			form.contacts = ''
			form.file_data = ''
			form.file_type = ''
			form.group_ids = []
			form.overwrite = 0
			form.default_active = 1
			form.status = 1
		}
	},
	onConfirm: async () => {
		await validateForm()
		await importSubscribers(toRaw(form))
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
