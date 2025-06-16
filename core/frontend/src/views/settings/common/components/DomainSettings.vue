<template>
	<n-form-item :label="$t('settings.common.network.domain')">
		<div class="flex-1 grid grid-cols-1 md:grid-cols-2 gap-4">
			<div>
				<n-input :value="currentDomain" disabled />
				<p class="mt-1 text-xs text-gray-500">
					{{ $t('settings.common.network.currentDomain') }}
				</p>
			</div>
			<div>
				<n-input
					v-model:value="networkForm.domain"
					:placeholder="$t('settings.common.network.newDomainPlaceholder')">
				</n-input>
				<p class="mt-1 text-xs text-gray-500">
					{{ $t('settings.common.network.domainRule') }}
				</p>
			</div>
		</div>
	</n-form-item>

	<div>
		<n-button type="primary" :disabled="!networkForm.domain" @click="handleUpdateDomain">
			{{ $t('settings.common.confirm') }}
		</n-button>
	</div>
</template>

<script lang="ts" setup>
import { setSystemConfigKey } from '@/api/modules/settings/common'
import { getSettingsStore } from '../store'

const { currentDomain, networkForm } = getSettingsStore()

const handleUpdateDomain = async () => {
	await setSystemConfigKey({ key: 'billionmail_hostname', value: networkForm.value.domain })
	// const newPath =
	// 	location.protocol + '//' + networkForm.value.domain + ':' + location.port + location.pathname
	// location.href = newPath
	currentDomain.value = networkForm.value.domain
	networkForm.value.domain = ''
}
</script>

<style lang="scss" scoped></style>
