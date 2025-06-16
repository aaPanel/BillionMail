<template>
	<n-form-item
		ref="formItemRef"
		:label="$t('settings.common.security.changePassword')"
		path="password">
		<div class="flex-1 flex flex-col gap-20px">
			<div>
				<n-input
					v-model:value="securityForm.newPassword"
					type="password"
					class="password-input"
					show-password-on="click"
					:placeholder="t('settings.common.security.newPassword')"
					@update:value="handleUpdatePassword">
				</n-input>
				<password-strength-indicator :strength="passwordStrength" />
			</div>
			<div>
				<n-input
					v-model:value="securityForm.confirmPassword"
					type="password"
					class="password-input"
					show-password-on="click"
					:placeholder="t('settings.common.security.confirmPassword')">
				</n-input>
			</div>
		</div>
	</n-form-item>
	<div>
		<n-button type="primary" :disabled="submitDisabled" @click="handleSubmitPassword">
			{{ t('settings.common.confirm') }}
		</n-button>
	</div>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { getSettingsStore } from '@/views/settings/common/store'
import { setSystemConfigKey } from '@/api/modules/settings/common'
import PasswordStrengthIndicator from './PasswordStrengthIndicator.vue'

const { t } = useI18n()
const { securityForm, passwordStrength, checkPasswordStrength } = getSettingsStore()

const formItemRef = useTemplateRef('formItemRef')

const rules: FormRules = {
	password: {
		trigger: ['input', 'blur'],
		validator: () => {
			if (!securityForm.value.newPassword || !securityForm.value.confirmPassword) {
				return true
			}
			if (securityForm.value.newPassword !== securityForm.value.confirmPassword) {
				return new Error(t('settings.common.security.passwordMismatch'))
			}
			return true
		},
	},
}

const submitDisabled = computed(() => {
	return !securityForm.value.newPassword || !securityForm.value.confirmPassword
})

const handleUpdatePassword = () => {
	checkPasswordStrength()
}

const handleSubmitPassword = async () => {
	await formItemRef.value?.validate()
	await setSystemConfigKey({ key: 'admin_password', value: securityForm.value.newPassword })
	securityForm.value.newPassword = ''
	securityForm.value.confirmPassword = ''
}

defineExpose({
	rules,
})
</script>
