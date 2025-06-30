<template>
	<div class="h-full p-20px">
		<n-breadcrumb class="mb-16px">
			<n-breadcrumb-item>
				<router-link to="/settings">{{ $t('layout.menu.settings') }}</router-link>
			</n-breadcrumb-item>
			<n-breadcrumb-item>
				<router-link to="/settings/service">{{ $t('layout.menu.service') }}</router-link>
			</n-breadcrumb-item>
			<n-breadcrumb-item>Rspamd</n-breadcrumb-item>
		</n-breadcrumb>
		<iframe class="w-full" :src="src"></iframe>
	</div>
</template>

<script lang="ts" setup>
import { isDev } from '@/utils'

const src = ref(`${window.location.origin}${isDev ? '/api' : ''}/rspamd/`)
</script>

<style lang="scss" scoped>
iframe {
	height: calc(100% - 30px);
}
</style>
