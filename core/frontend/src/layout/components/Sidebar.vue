<template>
	<n-layout-sider
		collapse-mode="width"
		:collapsed="isCollapse"
		:width="200"
		:collapsed-width="64"
		:content-style="{
			display: 'flex',
			flexDirection: 'column',
			height: '100%',
			overflow: 'hidden',
		}">
		<!-- 应用标志和名称 -->
		<div class="app-logo" :class="{ collapse: isCollapse }">
			<a href="/">
				<img class="icon" src="@/assets/images/logo.png"></img>
				<span v-show="!isCollapse" class="app-name">BillionMail</span>
			</a>
		</div>

		<!-- 导航菜单 -->
		<div class="nav-section">
			<n-menu
				:value="activeMenuKey"
				:collapsed="isCollapse"
				:collapsed-width="64"
				:options="menuOptions"
				:root-indent="24"
				@update:value="handleUpdateMenu">
			</n-menu>
		</div>
		<!-- 退出登录 -->
		<div class="footer-section">
			<n-menu
				value=""
				:collapsed="isCollapse"
				:collapsed-width="64"
				:options="logoutOptions"
				:root-indent="24"
				@update:value="handleUpdateMenu">
			</n-menu>
		</div>
	</n-layout-sider>
</template>

<script lang="tsx" setup>
import { VNodeChild } from 'vue'
import { MenuOption } from 'naive-ui'
import { storeToRefs } from 'pinia'
import { RouterLink } from 'vue-router'
import { useMenuStore, useGlobalStore, useUserStore } from '@/store'
import { menuList } from '@/router/router'

const { t } = useI18n()

const route = useRoute()

const menuStore = useMenuStore()
const userStore = useUserStore()
const globalStore = useGlobalStore()

const { isCollapse } = storeToRefs(globalStore)

// 当前菜单名称
const activeMenuKey = computed(() => {
	return String(route.meta?.key || '')
})

// 路由菜单
const routerMenus = computed(() => {
	return menuStore.menuList.filter(route => route.meta && !route.meta.hidden)
})

// 导航菜单选项
const menuOptions = computed(() => {
	return routerMenus.value.map(route => {
		const name = String(route.children?.[0]?.name || '')
		const key = String(route.meta?.key || '')
		const titleKey = String(route.meta?.titleKey || '')
		const title = titleKey ? t(titleKey) : String(route.meta?.title || '')
		return {
			key,
			label: () => renderLabel(name, title),
			icon: () => renderIcon(key),
		}
	})
})

const logoutOptions = ref<MenuOption[]>([
	{
		key: 'logout',
		label: () => <span class="ml-10px">{t('layout.menu.logout')}</span>,
		icon: () => renderIcon('logout'),
	},
])

const renderLabel = (name: string, title: string) => {
	return (
		<RouterLink class="flex items-center" to={{ name }}>
			<span>{title}</span>
		</RouterLink>
	)
}

const iconMap: Record<string, VNodeChild> = {
	overview: <i class="i-mdi-web"></i>,
	market: <i class="i-mdi-email-fast-outline"></i>,
	api: <i class="i-mdi-chart-line"></i>,
	contacts: <i class="i-mdi-user-multiple-outline"></i>,
	domain: <i class="i-mdi-web"></i>,
	mailbox: <i class="i-custom:mailbox"></i>,
	smtp: <i class="i-custom:smtp"></i>,
	settings: <i class="i-mdi-settings-outline"></i>,
	template: <i class="i-mdi-settings-outline"></i>,
	rspamd: <i class="i-mdi-settings-outline"></i>,
	logout: <i class="i-mdi-logout"></i>,
}

const renderIcon = (key: string) => {
	return iconMap[key]
}

const handleUpdateMenu = (key: string) => {
	if (key === 'logout') {
		userStore.logout()
	}
	if (key === 'webmail') {
		const route = routerMenus.value.find(item => item.meta?.key === 'webmail')
		if (route) {
			const href = String(route.meta?.href)
			window.open(href)
		}
	}
}

onMounted(() => {
	menuStore.setMenuList(menuList)
})
</script>

<style lang="scss" scoped>
.n-layout-sider {
	box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
	z-index: 1010;
}

.app-logo {
	display: flex;
	padding: 16px 24px;
	border-bottom: 1px solid var(--color-border-1);
	transition: all 0.3s ease;

	&.collapse {
		justify-content: center;
		padding: 16px 0;
	}

	a {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.icon {
		width: 36px;
	}
	
	.app-name {
		font-size: 19px;
		font-weight: bold;
		color: var(--color-text-5);
	}
}

.nav-section {
	flex: 1;
	overflow: auto;
}

.footer-section {
	border-top: 1px solid var(--color-border-1);
	transition: all 0.3s ease;
}

.n-menu {
	--n-item-height: 48px;
	--n-font-size: 14px;
	padding: 16px 0;

	:deep(.n-menu-item) {
		margin-top: 0;
		margin-bottom: 8px;

		&:last-of-type {
			margin-bottom: 0;
		}

		.n-menu-item-content {
			padding-right: 24px;
			line-height: 24px;
		}
	}
}
</style>
