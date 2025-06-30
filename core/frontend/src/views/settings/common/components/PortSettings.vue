<template>
	<div>
		<label class="block text-sm font-medium text-basic mb-3">
			{{ t('settings.common.system.port') }}
		</label>
		<div class="grid items-start grid-cols-1 md:grid-cols-2 gap-4">
			<div>
				<n-input v-model:value="port" @update:value="handleUpdatePort"></n-input>
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
			<command :value="commandRef"></command>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { getSettingsStore } from '@/views/settings/common/store'
import Command from './Command.vue'

const { t } = useI18n()

const { currentPort, portCommand } = getSettingsStore()

const port = ref(currentPort.value)

const commandRef = ref('')

const handleUpdatePort = (val: string) => {
	commandRef.value = portCommand.value.replace(/PORT/g, val)
}

watchEffect(() => {
	port.value = currentPort.value
	// 初始化命令，替换 PORT 占位符
	if (currentPort.value) {
		commandRef.value = portCommand.value.replace(/PORT/g, currentPort.value)
	}
})
</script>

<style lang="scss" scoped></style>
