<template>
	<n-card class="mb-8" :bordered="false">
		<template #header>
			<div class="flex items-center">
				<n-icon class="mr-2 text-primary" size="20">
					<i class="i-mdi-network"></i>
				</n-icon>
				<span class="text-lg font-medium">{{ $t('settings.common.network.title') }}</span>
			</div>
		</template>

		<n-form>
			<SSLSettings />
			<n-divider />
			<DomainSettings />
		</n-form>
	</n-card>
</template>

<script lang="ts" setup>
import SSLSettings from './SSLSettings.vue'
import DomainSettings from './DomainSettings.vue'
</script>
