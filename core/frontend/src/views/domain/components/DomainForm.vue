<template>
	<modal :title="title" :width="540">
		<div class="pt-16px">
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-form-item :label="t('domain.form.domain')" path="domain">
					<n-input
						v-model:value="form.domain"
						:disabled="isEdit"
						:placeholder="t('domain.form.domainPlaceholder')">
					</n-input>
				</n-form-item>
				<n-form-item v-if="false" label="A记录">
					<n-input
						v-model:value="form.a_record"
						:disabled="isEdit"
						placeholder="请输入A记录，例如：mail.aapanel.com">
					</n-input>
				</n-form-item>
				<n-form-item :label="t('domain.form.quota')">
					<n-input-number v-model:value="form.quota" class="flex-1" :min="0" :show-button="false">
					</n-input-number>
					<div class="w-100px ml-20px">
						<n-select v-model:value="form.quota_unit" :options="unitOptions"></n-select>
					</div>
				</n-form-item>
				<n-form-item :label="t('domain.form.mailboxCount')">
					<n-input-number
						v-model:value="form.mailboxes"
						class="flex-1"
						:min="0"
						:show-button="false">
					</n-input-number>
				</n-form-item>
				<n-form-item :label="t('domain.form.globalCatch')">
					<n-input
						v-model:value="form.email"
						:placeholder="t('domain.form.globalCatchPlaceholder')">
					</n-input>
				</n-form-item>
			</bt-form>
			<bt-tips>
				<li class="text-error">{{ t('domain.form.tips.aRecordFailed') }}</li>
				<li>{{ t('domain.form.tips.dnsSetup') }}</li>
				<li>{{ t('domain.form.tips.cloudflare') }}</li>
			</bt-tips>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { getByteUnit, getNumber } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { createDomain, updateDomain } from '@/api/modules/domain'
import type { MailDomain } from '../interface'

const { t } = useI18n()

const isEdit = ref(false)

const title = computed(() => {
	return isEdit.value ? t('domain.form.editTitle') : t('domain.form.addTitle')
})

const formRef = useTemplateRef('formRef')

const form = reactive({
	domain: '',
	a_record: '',
	quota: 5,
	quota_unit: 'GB',
	mailboxes: 50,
	email: '',
})

const unitOptions = [
	{ label: 'GB', value: 'GB' },
	{ label: 'MB', value: 'MB' },
]

const rules: FormRules = {
	domain: {
		trigger: ['blur', 'input'],
		validator: () => {
			if (form.domain.trim() === '') {
				return new Error(t('domain.form.validation.domainRequired'))
			}
			return true
		},
	},
}

/**
 * @description Calculate the byte number based on the domain quota and unit
 * @param quota
 * @param quota_unit
 */
const getQuotaByte = (quota: number, quota_unit: string) => {
	switch (quota_unit) {
		case 'GB':
			return quota * 1024 * 1024 * 1024
		case 'MB':
			return quota * 1024 * 1024
		default:
			return quota
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ isEdit: boolean; row: MailDomain | null }>()
			isEdit.value = state.isEdit
			if (state.row) {
				const { row } = state
				form.domain = row.domain
				form.a_record = row.a_record
				const quota = getByteUnit(row.quota)
				form.quota = getNumber(quota.split(' ')[0])
				form.quota_unit = quota.split(' ')[1]
				form.mailboxes = row.mailboxes
				form.email = row.email
			}
		} else {
			form.domain = ''
			form.a_record = ''
			form.quota = 5
			form.quota_unit = 'GB'
			form.mailboxes = 50
			form.email = ''
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()
		if (isEdit.value) {
			await updateDomain({
				domain: form.domain,
				quota: getQuotaByte(form.quota, form.quota_unit),
				mailboxes: form.mailboxes,
				email: form.email,
			})
		} else {
			await createDomain({
				domain: form.domain,
				quota: getQuotaByte(form.quota, form.quota_unit),
				mailboxes: form.mailboxes,
				email: form.email,
			})
		}
		const state = modalApi.getState<{ refresh: Function }>()
		state.refresh()
	},
})
</script>
