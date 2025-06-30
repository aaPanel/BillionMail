import type { RouteRecordRaw } from 'vue-router'
import { isDev } from '@/utils'

export default {
	path: '/template',
	meta: {
		sort: 3,
		title: 'template',
		key: 'template',
		titleKey: 'layout.menu.template',
		hidden: isDev,
	},
	component: () => import('@/views/template/index.vue'),
} as RouteRecordRaw
