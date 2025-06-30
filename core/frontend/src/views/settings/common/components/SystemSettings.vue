<template>
	<n-card class="mb-8" :bordered="false">
		<template #header>
			<div class="flex items-center">
				<n-icon class="mr-2 text-primary" size="20">
					<i class="i-mdi-cog"></i>
				</n-icon>
				<span class="text-lg font-medium">{{ $t('settings.common.system.title') }}</span>
			</div>
		</template>

		<div class="space-y-8">
			<PortSettings />
			<n-divider />
			<TimezoneSettings />
			<n-divider />
			<IpWhitelistSettings />
		</div>
	</n-card>
</template>

<script lang="ts" setup>
import PortSettings from './PortSettings.vue'
import TimezoneSettings from './TimezoneSettings.vue'
import IpWhitelistSettings from './IpWhitelistSettings.vue'
</script>
