import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/domain',
	component: Layout,
	meta: { sort: 5, key: 'domain', title: 'MailDomain', titleKey: 'layout.menu.domain' },
	children: [
		{
			path: '/domain',
			name: 'Domain',
			component: () => import('@/views/domain/index.vue'),
		},
		{
			path: 'edit-domain',
			name: 'EditDomain',
			component: () => import('@/views/domain/pages/editDomain/index.vue'),
		},
	],
}

export default route
