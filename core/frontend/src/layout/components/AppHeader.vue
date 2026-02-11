<template>
	<n-layout-header ref="headerRef" :style="{ top: `${top}px` }">
		<div class="header-left">
			<n-button class="icon-btn" :bordered="false" @click="handleCollapse">
				<i class="icon" :class="isCollapse ? 'i-mdi-menu-close' : 'i-mdi-menu-open'"></i>
			</n-button>
			<n-button type="primary" text class="text-14px" @click="handleGoIssues">
				{{ t('layout.header.submit') }}
				<i class="i-mdi:arrow-right ml-1px"></i>
			</n-button>
		</div>

		<div class="header-right">
			<n-button class="icon-btn" :bordered="false" @click="handleSetTheme">
				<i class="icon" :class="theme === 'light' ? 'i-ri-sun-line' : 'i-ri-moon-line'"></i>
			</n-button>
			<n-dropdown
				v-if="langOptions.length > 0"
				size="large"
				:options="langOptions"
				@select="handleLangAction">
				<n-button class="icon-btn" :bordered="false">
					<i class="icon i-mdi-language text-20px"></i>
				</n-button>
			</n-dropdown>
			<InstanceSwitcher />
			<n-dropdown size="large" :options="userOptions" @select="handleUserAction">
				<n-button class="icon-btn" :bordered="false">
					<i class="icon i-mdi-user-outline"></i>
				</n-button>
			</n-dropdown>
			<n-button type="primary" text class="text-14px" @click="handleGoVersion">
				{{ version }}
			</n-button>
		</div>
	</n-layout-header>
</template>

<script lang="ts" setup>
import { storeToRefs } from 'pinia'
import { DropdownOption } from 'naive-ui'
import { useUserStore, useGlobalStore, useThemeStore } from '@/store'
import InstanceSwitcher from './InstanceSwitcher.vue'
import { getVersionInfo } from '@/api/modules/settings'
import { isObject } from '@/utils'

defineProps({
	top: {
		type: Number,
		default: 0,
	},
})

const { t } = useI18n()

const version = ref('--')

const userStore = useUserStore()

const globalStore = useGlobalStore()
const { isCollapse, langList } = storeToRefs(globalStore)

const themeStore = useThemeStore()
const { theme } = storeToRefs(themeStore)

const handleCollapse = () => {
	globalStore.setCollapse()
}

const handleGoIssues = () => {
	window.open('https://github.com/aaPanel/BillionMail/issues')
}

const langOptions = ref<DropdownOption[]>([])

const userOptions = ref<DropdownOption[]>([
	{
		label: t('layout.menu.logout'),
		key: 'logout',
	},
])

const handleSetTheme = () => {
	themeStore.setTheme(theme.value === 'dark' ? 'light' : 'dark')
}

const handleLangAction = async (key: string) => {
	await globalStore.setLang(key)
	window.location.reload()
}

const handleUserAction = (key: string) => {
	switch (key) {
		case 'logout':
			userStore.logout()
			break
	}
}

const handleGoVersion = () => {
	window.open('https://github.com/aaPanel/BillionMail/releases')
}

const getLangOptions = async () => {
	langOptions.value = langList.value.map(item => {
		return {
			label: item.cn,
			key: item.name,
		}
	})
}

const getVersion = async () => {
	const res = await getVersionInfo()
	if (isObject<{ version: string }>(res)) {
		version.value = `v${res.version}`
	}
}

onMounted(() => {
	getVersion()
	getLangOptions()
})
</script>

<style lang="scss" scoped>
.n-layout-header {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	display: flex;
	justify-content: space-between;
	align-items: center;
	height: 48px;
	padding: 0 20px 0 12px;
	box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
	z-index: 1000;
}

.header-left,
.header-right {
	display: flex;
	align-items: center;
	gap: 8px;
}

.header-item {
	color: var(--color-text-4);
	font-size: 14px;
	text-align: center;
	font-weight: 600;
}

.icon-btn {
	--n-width: 42px;
	--n-height: 48px;
	--n-padding: 0;
	--n-font-size: 22px;
	--n-text-color: var(--color-text-4);
	--n-ripple-color: none;
}
</style>
