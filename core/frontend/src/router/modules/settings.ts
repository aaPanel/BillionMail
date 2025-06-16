import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/settings',
	name: 'SettingsLayout',
	redirect: '/settings/common',
	meta: { sort: 9, key: 'settings', title: 'Settings', titleKey: 'layout.menu.settings' },
	component: Layout,
	children: [
		{
			path: '/settings',
			name: 'Settings',
			redirect: '/settings/common',
			component: () => import('@/views/settings/index.vue'),
			children: [
				{
					path: 'common',
					name: 'SettingsCommon',
					meta: { title: 'Common', titleKey: 'layout.menu.common' },
					component: () => import('@/views/settings/common/index.vue'),
				},
				{
					path: 'service',
					name: 'SettingsService',
					meta: { title: 'Service', titleKey: '' },
					component: () => import('@/views/settings/service/index.vue'),
				},
				{
					path: 'bcc',
					name: 'SettingsBcc',
					meta: { title: 'BCC', titleKey: 'layout.menu.bcc' },
					component: () => import('@/views/settings/bcc/index.vue'),
				},
				{
					path: 'forward',
					name: 'SettingsForward',
					meta: { title: 'Forward', titleKey: 'layout.menu.forward' },
					component: () => import('@/views/settings/forward/index.vue'),
				},
			],
		},
	],
}

export default route
