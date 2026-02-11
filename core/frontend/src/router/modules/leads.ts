import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/leads',
	name: 'LeadsLayout',
	component: Layout,
	meta: {
		sort: 6,
		key: 'leads',
		title: 'Leads',
		titleKey: 'layout.menu.leads',
	},
	children: [
		{
			path: '/leads',
			name: 'Leads',
			meta: { title: 'Leads', titleKey: 'layout.menu.leads' },
			component: () => import('@/views/leads/index.vue'),
		},
	],
}

export default route
