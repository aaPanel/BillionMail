<template>
	<modal :title="title" :width="1200">
		<bt-form class="pt-8px" :model="form">
			<n-grid :cols="24" :x-gap="24">
				<n-form-item-gi :span="18" label="Name" path="temp_name">
					<n-input
						v-model:value="form.temp_name"
						:placeholder="$t('market.template.add.namePlaceholder')">
					</n-input>
				</n-form-item-gi>
				<n-form-item-gi :span="6" label="Type" path="add_type">
					<n-select v-model:value="form.add_type" :disabled="isEdit" :options="typeOptions">
					</n-select>
				</n-form-item-gi>
			</n-grid>
			<n-form-item :label="typeLabel" :show-feedback="false">
				<div class="flex flex-col flex-1 h-580px">
					<div class="flex justify-end gap-12px mb-8px">
						<div>
							<bt-file-upload
								mode="button"
								button-type="primary"
								:is-upload="false"
								:accept="['html']"
								@change="handleFileUpload">
								{{ $t('market.template.uploadHtml') }}
							</bt-file-upload>
						</div>
						<n-button type="primary" ghost @click="handlePreview">
							{{ $t('common.actions.preview') }}
						</n-button>
					</div>
					<div class="flex-1">
						<template v-if="form.add_type === 1">
							<email-editor v-model:config="form.drag_data" v-model:html="form.html_content">
							</email-editor>
						</template>
						<template v-if="form.add_type === 0">
							<bt-editor v-model:value="form.html_content" language="html"> </bt-editor>
						</template>
					</div>
				</div>
			</n-form-item>
		</bt-form>
		<template #footer-left> </template>
	</modal>
</template>

<script lang="ts" setup>
import { useModal } from '@/hooks/modal/useModal'
import { addTemplate, updateTemplate } from '@/api/modules/market/template'
import type { Template } from '../interface'
import { UploadFileInfo } from 'naive-ui'

const EmailEditor = defineAsyncComponent(() => import('@/features/EmailEditor/index.vue'))

const isEdit = ref(false)

const title = computed(() => {
	return isEdit.value ? 'Edit template' : 'New template'
})

const id = ref(0)

const form = reactive({
	temp_name: '',
	add_type: 0,
	html_content: '',
	drag_data: '',
})

const typeOptions = [
	{ label: 'HTML', value: 0 },
	{ label: 'Drag', value: 1 },
]

const typeLabel = computed(() => {
	return typeOptions.find(item => item.value === form.add_type)?.label || ''
})

const handleFileUpload = (file: UploadFileInfo) => {
	const reader = new FileReader()
	reader.onload = e => {
		form.html_content = e.target?.result as string
	}
	if (file.file) {
		reader.readAsText(file.file)
	}
}

const handlePreview = () => {
	// eslint-disable-next-line no-unused-vars
	const state = modalApi.getState<{ preview: (content: string) => void }>()
	if (state.preview) {
		state.preview(form.html_content)
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ isEdit: boolean; row: Template }>()
			const { row } = state
			isEdit.value = state.isEdit
			if (row) {
				id.value = row.id
				form.temp_name = row.temp_name
				form.add_type = row.add_type
				form.html_content = row.html_content
				form.drag_data = row.drag_data
			}
		} else {
			id.value = 0
			form.temp_name = ''
			form.add_type = 0
			form.html_content = ''
			form.drag_data = ''
		}
	},
	onConfirm: async () => {
		if (isEdit.value) {
			await updateTemplate({ id: id.value, ...toRaw(form) })
		} else {
			await addTemplate(toRaw(form))
		}
		const state = modalApi.getState<{ refresh: () => void }>()
		state.refresh()
	},
})
</script>

<style lang="scss" scoped></style>
