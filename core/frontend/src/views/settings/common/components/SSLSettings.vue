<template>
	<div>
		<label class="block text-sm font-medium mb-3">
			{{ t('settings.common.network.ssl') }}
		</label>

		<div class="flex items-center mb-4">
			<n-switch v-model:value="networkForm.sslEnabled" disabled />
			<label class="ml-2 font-medium">
				{{ t('settings.common.network.sslEnabled') }}
			</label>
		</div>

		<div class="mt-6">
			<label class="block text-sm font-medium mb-2">
				{{ t('settings.common.network.certStatus') }}
			</label>
			<div class="flex items-center">
				<div
					class="w-3 h-3 rounded-full mr-2 animate-pulse"
					:class="sslInfo.status ? 'bg-primary' : 'bg-error'"></div>
				<span class="text-sm">
					{{
						sslInfo.status
							? t('settings.common.network.certValidStatus')
							: t('settings.common.network.certExpiredStatus')
					}}，
				</span>
				<span class="text-sm">
					{{ t('settings.common.network.certExpireTime', { time: sslInfo.expireTime }) }}
				</span>
				<span class="text-sm">
					{{ t('settings.common.network.certValid', { days: sslInfo.expireDays }) }}
				</span>
			</div>
		</div>

		<div class="mt-6">
			<label class="block text-sm font-medium mb-2">
				{{ t('settings.common.network.sslCert') }}
			</label>
			<n-input
				v-model:value="networkForm.sslCert"
				type="textarea"
				:rows="6"
				:placeholder="t('settings.common.network.certPlaceholder')"
				class="font-mono text-sm">
			</n-input>
		</div>

		<div class="mt-4">
			<label class="block text-sm font-medium mb-2">
				{{ t('settings.common.network.sslKey') }}
			</label>
			<n-input
				v-model:value="networkForm.sslKey"
				type="textarea"
				:rows="6"
				:placeholder="t('settings.common.network.keyPlaceholder')"
				class="font-mono text-sm">
			</n-input>
		</div>

		<div class="mt-6">
			<n-button type="primary" :disabled="updateBtnDisabled" @click="handleUpdateSSL">
				{{ t('settings.common.network.updateSSL') }}
			</n-button>
			<n-button type="primary" class="ml-12px" ghost @click="handleApplySSL">
				Apply Certificate
			</n-button>
		</div>
	</div>
</template>

<script lang="tsx" setup>
import { applyCert, setSslConfig } from '@/api/modules/settings/common'
import { getSettingsStore } from '../store'
import { confirm } from '@/utils'

const { t } = useI18n()

const { networkForm, sslInfo, currentDomain, serverIp, getCommonConfig } = getSettingsStore()

const updateBtnDisabled = computed(() => {
	return !networkForm.value.sslCert || !networkForm.value.sslKey
})

const handleUpdateSSL = async () => {
	await setSslConfig({
		certPem: networkForm.value.sslCert,
		privateKey: networkForm.value.sslKey,
	})
	getCommonConfig()
}

const handleApplySSL = () => {
	confirm({
		title: 'Apply SSL',
		width: 460,
		content: () => (
			<div style={{ lineHeight: '22px' }}>
				<div class="mb-8px">Certificate Application Functionality:</div>
				<div>- Implement certificate application using hostname</div>
				<div>- Ensure the hostname has a valid A record configured with:</div>
				<div>
					- Record Type: <b>A</b>
				</div>
				<div>
					- Host: <b>{currentDomain.value}</b>
				</div>
				<div>
					- Value: <b>{serverIp.value}</b>
				</div>
			</div>
		),
		onConfirm: async () => {
			await applyCert()
			getCommonConfig()
		},
	})
}
</script>

<style lang="scss" scoped></style>
