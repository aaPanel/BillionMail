<template>
	<div>
		<label class="block text-sm font-medium text-gray-700 mb-3">
			{{ t('settings.common.system.port') }}
		</label>
		<div class="grid items-start grid-cols-1 md:grid-cols-2 gap-4">
			<div>
				<n-input :value="currentPort" disabled />
				<p class="mt-1 text-xs text-desc">{{ t('settings.common.system.currentPort') }}</p>
			</div>
			<div class="flex items-center">
				<n-alert type="warning" class="w-full">
					<template #icon>
						<n-icon><i class="i-mdi-alert-circle" /></n-icon>
					</template>
					{{ t('settings.common.system.terminalRequired') }}
				</n-alert>
			</div>
		</div>

		<div class="mt-5">
			<command :value="portCommand"></command>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { getSettingsStore } from '@/views/settings/common/store'
import Command from './Command.vue'

const { t } = useI18n()

const { currentPort, portCommand } = getSettingsStore()
</script>

<style lang="scss" scoped></style>
