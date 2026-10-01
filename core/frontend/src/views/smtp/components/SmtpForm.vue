<template>
	<modal :title="title" width="500">
		<bt-form ref="formRef" :model="form" :rules="rules" class="pt-16px">
			<n-grid :cols="1" :x-gap="24">
				<n-form-item-gi :span="1" :label="$t('smtp.form.remark')" path="remark">
					<n-input v-model:value="form.remark"></n-input>
				</n-form-item-gi>
			</n-grid>
			<n-form-item :label="$t('smtp.form.domain')" :path="isEdit ? '' : 'sender_domains'">
				<n-select
					v-model:value="form.sender_domains"
					multiple
					max-tag-count="responsive"
					:options="domainOptions">
				</n-select>
			</n-form-item>
			<n-grid :cols="8" :x-gap="24">
				<n-form-item-gi :span="6" :label="$t('smtp.form.server')" path="relay_host">
					<n-input v-model:value="form.relay_host"></n-input>
				</n-form-item-gi>
				<n-form-item-gi :span="2" :label="$t('smtp.form.port')" path="relay_port">
					<n-input-number
						v-model:value="form.relay_port"
						class="flex-1"
						:show-button="false"
						:min="1">
					</n-input-number>
				</n-form-item-gi>
			</n-grid>
			<n-form-item :label="$t('smtp.form.username')" path="auth_user">
				<n-input v-model:value="form.auth_user"></n-input>
			</n-form-item>
			<n-form-item :label="$t('smtp.form.password')" path="auth_password">
				<n-input v-model:value="form.auth_password" class="password-input" type="password">
				</n-input>
			</n-form-item>
			<bt-more :title="$t('smtp.form.advancedFeatures')">
				<n-form-item :label="$t('smtp.form.authMethod')" path="auth_method">
					<n-select v-model:value="form.auth_method" :options="authMethodOptions"> </n-select>
				</n-form-item>
				<n-form-item :label="$t('smtp.form.skipTlsVerify')" path="skip_tls_verify">
					<n-select v-model:value="form.skip_tls_verify" :options="tlsOptions"> </n-select>
				</n-form-item>
				<n-form-item label="HELO Name" path="helo_name">
					<n-input v-model:value="form.helo_name"> </n-input>
				</n-form-item>
			</bt-more>
		</bt-form>
		<template #footer-left>
			<n-button type="primary" ghost @click="onTest">{{ $t('smtp.form.testConnection') }}</n-button>
		</template>
	</modal>
</template>

<script lang="ts" setup>
import { FormRules, SelectOption } from 'naive-ui'
import { getNumber, isArray } from '@/utils'
import { addSmtp, editSmtp, getUnbindDomains, testSmtp } from '@/api/modules/smtp'
import { useModal } from '@/hooks/modal/useModal'
import type { SmtpService } from '../types/base'

const { t } = useI18n()

const isEdit = ref(false)

const title = computed(() => {
	return isEdit.value ? t('smtp.form.editTitle') : t('smtp.form.addTitle')
})

const id = ref(0)

const formRef = useTemplateRef('formRef')

const form = reactive({
	rtype: 'custom',
	smtp_name: '',
	sender_domains: [] as string[],
	relay_host: '',
	relay_port: 587,
	auth_user: '',
	auth_password: '',
	remark: '',
	active: 1,
	auth_method: 'LOGIN' as string | null,
	skip_tls_verify: 0 as number | null,
	helo_name: '',
})

const rules: FormRules = {
	smtp_name: {
		required: true,
		message: t('smtp.form.validation.nameRequired'),
		trigger: ['blur', 'input'],
	},
	relay_host: {
		required: true,
		message: t('smtp.form.validation.serverRequired'),
		trigger: ['blur', 'input'],
	},
	relay_port: {
		required: true,
		trigger: ['blur', 'input'],
		validator: (rule, value) => {
			if (!value) {
				return new Error(t('smtp.form.validation.portRequired'))
			}
			return true
		},
	},
}

const domainOptions = ref<SelectOption[]>([])

const authMethodOptions: SelectOption[] = [
	{ label: 'LOGIN', value: 'LOGIN' },
	{ label: 'PLAIN', value: 'PLAIN' },
	{ label: 'CRAM-MD5', value: 'CRAM-MD5' },
	{ label: 'NONE', value: 'NONE' },
]

const tlsOptions: SelectOption[] = [
	{ label: t('smtp.form.tlsOptions.skip'), value: 1 },
	{ label: t('smtp.form.tlsOptions.noSkip'), value: 0 },
]

const getDomainList = async () => {
	const res = await getUnbindDomains()
	if (isArray<{ domain: string; is_bound: boolean }>(res)) {
		domainOptions.value = res.map(item => {
			return {
				label: `${item.domain}${!form.sender_domains.includes(item.domain) && item.is_bound ? t('smtp.form.bind') : ''}`,
				value: item.domain,
				disabled: !form.sender_domains.includes(item.domain) && item.is_bound,
			}
		})
	}
}

const typeMap = {
	gmail: {
		smtp_name: 'Gmail',
		relay_host: 'smtp.gmail.com',
		relay_port: 587,
		username: '',
		password: '',
		remark: 'Gmail',
	},
	mailgun: {
		smtp_name: 'Mailgun',
		relay_host: 'smtp.mailgun.org',
		relay_port: 587,
		username: '',
		password: '',
		remark: 'Mailgun',
	},
	aws: {
		smtp_name: 'AWS SES',
		relay_host: 'email-smtp.us-east-1.amazonaws.com',
		relay_port: 587,
		username: '',
		password: '',
		remark: 'AWS SES',
	},
	sendgrid: {
		smtp_name: 'SendGrid',
		relay_host: 'smtp.sendgrid.net',
		relay_port: 587,
		username: 'apikey',
		password: '',
		remark: 'SendGrid',
	},
	custom: {
		smtp_name: 'Custom SMTP Relay',
		relay_host: '',
		relay_port: 587,
		username: '',
		password: '',
		remark: 'Custom SMTP Relay',
	},
}

const generateForm = () => {
	const data = typeMap[form.rtype as keyof typeof typeMap]
	if (data) {
		form.smtp_name = data.smtp_name
		form.relay_host = data.relay_host
		form.relay_port = data.relay_port
		form.auth_user = data.username
		form.auth_password = data.password
		form.remark = data.remark
	}
}

const resetForm = () => {
	id.value = 0

	form.smtp_name = ''
	form.sender_domains = []
	form.relay_host = ''
	form.relay_port = 587
	form.auth_user = ''
	form.auth_password = ''
	form.remark = ''
	form.active = 1
	form.skip_tls_verify = null
	form.auth_method = null
	form.helo_name = ''
}

const onTest = async () => {
	await formRef.value?.validate()
	await testSmtp({
		sender_domains: form.sender_domains,
		relay_host: form.relay_host,
		relay_port: form.relay_port,
		auth_user: form.auth_user,
		auth_password: form.auth_password,
	})
}

const getParams = () => {
	return {
		rtype: form.rtype,
		smtp_name: form.smtp_name,
		sender_domains: form.sender_domains,
		relay_host: form.relay_host,
		relay_port: form.relay_port,
		auth_user: form.auth_user,
		auth_password: form.auth_password,
		remark: form.remark,
		active: form.active,
		skip_tls_verify: form.skip_tls_verify || '',
		auth_method: form.auth_method || '',
		helo_name: form.helo_name,
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			getDomainList()

			const state = modalApi.getState<{ isEdit: boolean; type: string; row: SmtpService | null }>()
			form.rtype = state.type
			isEdit.value = state.isEdit
			generateForm()

			if (state.row) {
				const { row } = state
				id.value = row.id
				form.rtype = row.rtype
				form.smtp_name = row.smtp_name
				form.sender_domains = row.sender_domains
				form.relay_host = row.relay_host
				form.relay_port = getNumber(row.relay_port)
				form.auth_user = row.auth_user
				form.auth_password = row.auth_password
				form.remark = row.remark
				form.active = row.active
				form.skip_tls_verify = row.skip_tls_verify
			}
		} else {
			resetForm()
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()
		const params = getParams()
		if (isEdit.value) {
			await editSmtp({ ...params, id: id.value })
		} else {
			await addSmtp(params)
		}
		const state = modalApi.getState<{ refresh: Function }>()
		state.refresh()
	},
})
</script>

<style lang="scss" scoped></style>
