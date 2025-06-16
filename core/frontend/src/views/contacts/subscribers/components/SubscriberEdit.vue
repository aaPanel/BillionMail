<template>
	<modal :title="$t('contacts.subscribers.edit.title')" width="600">
		<div class="pt-12px">
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-grid :cols="12" :x-gap="24">
					<n-form-item-gi :span="8" :label="$t('contacts.subscribers.edit.email')">
						<n-input v-model:value="form.email" :disabled="true"></n-input>
					</n-form-item-gi>
					<n-form-item-gi :span="4" label="Status">
						<n-select v-model:value="form.active" :options="activeOptions"></n-select>
					</n-form-item-gi>
				</n-grid>
				<n-form-item :label="$t('contacts.subscribers.edit.group')" path="group_ids">
					<group-select v-model:value="form.group_ids"></group-select>
				</n-form-item>
				<n-form-item label="Attributes" path="group_ids">
					<n-input v-model:value="form.attribs" type="textarea" :rows="6"></n-input>
				</n-form-item>
			</bt-form>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { editContact } from '@/api/modules/contacts/subscribers'
import type { Subscriber } from '../interface'

import GroupSelect from './GroupMultipleSelect.vue'
import { Message } from '@/utils'

const { t } = useI18n()

const formRef = useTemplateRef('formRef')

const form = reactive({
	email: '',
	active: 1,
	attribs: '',
	group_ids: [] as number[],
})

const rules: FormRules = {
	group_ids: {
		trigger: 'change',
		validator: () => {
			if (form.group_ids.length === 0) {
				return new Error(t('contacts.subscribers.edit.validation.groupRequired'))
			}
			return true
		},
	},
}

const activeOptions = [
	{ label: 'Subscribe', value: 1 },
	{ label: 'Unsubscribe', value: 0 },
]

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ row: Subscriber | null }>()
			const { row } = state
			if (row) {
				form.email = row.email
				form.active = row.active
				form.group_ids = row.groups.map(item => item.id)
				form.attribs = row.attribs ? JSON.stringify(row.attribs, null, 2) : ''
			}
		} else {
			form.email = ''
			form.active = 1
			form.group_ids = []
			form.attribs = ''
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()

		let attribs = {}
		try {
			attribs = form.attribs ? JSON.parse(form.attribs) : {}
		} catch (error) {
			Message.error(`${error}`, { close: true })
			return false
		}
		await editContact({
			emails: form.email,
			group_ids: form.group_ids,
			active: form.active,
			attribs: JSON.stringify(attribs),
		})
		const state = modalApi.getState<{ refresh: () => void }>()
		state.refresh()
	},
})
</script>
