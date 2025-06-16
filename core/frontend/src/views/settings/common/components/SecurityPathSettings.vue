<template>
	<n-form-item :label="t('settings.common.security.securityPath')">
		<div class="flex-1 grid grid-cols-1 md:grid-cols-2 gap-4">
			<div>
				<n-input :value="currentPath" disabled />
				<p class="mt-1 text-xs text-desc">
					{{ t('settings.common.security.currentPath') }}
				</p>
			</div>
			<div>
				<n-input
					v-model:value="securityForm.securityPath"
					:placeholder="t('settings.common.security.newPathPlaceholder')">
				</n-input>
				<p class="mt-1 text-xs text-desc">
					{{ t('settings.common.security.pathRule') }}
				</p>
			</div>
		</div>
	</n-form-item>
	<div class="mt-4px">
		<n-button
			type="primary"
			:disabled="!securityForm.securityPath"
			@click="handleSubmitSecurityPath">
			{{ t('settings.common.confirm') }}
		</n-button>
	</div>
</template>

<script lang="ts" setup>
import { setSystemConfigKey } from '@/api/modules/settings/common'
import { getSettingsStore } from '@/views/settings/common/store'

const { t } = useI18n()
const { currentPath, securityForm } = getSettingsStore()

const handleSubmitSecurityPath = async () => {
	await setSystemConfigKey({ key: 'safe_path', value: securityForm.value.securityPath })
	currentPath.value = securityForm.value.securityPath
	securityForm.value.securityPath = ''
}
</script>
