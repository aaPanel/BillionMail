<template>
	<n-card class="mb-8" :bordered="false">
		<template #header>
			<div class="flex items-center">
				<n-icon class="mr-2 text-primary" size="20">
					<i class="i-mdi-shield-account"></i>
				</n-icon>
				<span class="text-lg font-medium">{{ t('settings.common.security.title') }}</span>
			</div>
		</template>

		<n-form :model="securityForm" :rules="rules">
			<UsernameSettings ref="usernameRef" />
			<n-divider />
			<PasswordSettings ref="passwordRef" />
			<n-divider />
			<SecurityPathSettings />
		</n-form>
	</n-card>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { getSettingsStore } from '@/views/settings/common/store'

import UsernameSettings from './UsernameSettings.vue'
import PasswordSettings from './PasswordSettings.vue'
import SecurityPathSettings from './SecurityPathSettings.vue'

const { t } = useI18n()

const { securityForm } = getSettingsStore()

const rules = ref<FormRules>({})

const usernameRef = useTemplateRef('usernameRef')

const passwordRef = useTemplateRef('passwordRef')

onMounted(() => {
	if (usernameRef.value) {
		rules.value = {
			...rules.value,
			...usernameRef.value.rules,
		}
	}
	if (passwordRef.value) {
		rules.value = {
			...rules.value,
			...passwordRef.value.rules,
		}
	}
})
</script>
