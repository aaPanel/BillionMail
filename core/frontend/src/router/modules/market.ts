import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/market',
	redirect: '/market/task',
	name: 'MarketLayout',
	component: Layout,
	meta: {
		sort: 2,
		key: 'market',
		title: 'Email Marketing',
		titleKey: 'layout.menu.market',
	},
	children: [
		{
			path: '/market',
			redirect: '/market/task',
			name: 'Market',
			component: () => import('@/views/market/index.vue'),
			children: [
				{
					path: 'task',
					name: 'MarketTask',
					meta: { title: 'Task', titleKey: 'layout.menu.task' },
					component: () => import('@/views/market/task/index.vue'),
				},
				{
					path: 'template',
					name: 'MarketTemplate',
					meta: { title: 'Template', titleKey: 'layout.menu.template' },
					component: () => import('@/views/market/template/index.vue'),
				},
			],
		},
		{
			path: 'template/edit/:id',
			name: 'MarketTemplateEdit',
			meta: { title: 'Template', hidden: true },
			component: () => import('@/views/market/template/edit.vue'),
		},
		{
			path: 'task/edit',
			name: 'MarketTaskEdit',
			meta: { title: 'Task', hidden: true },
			component: () => import('@/views/market/task/edit.vue'),
		},
		{
			path: 'task/analytics/:id',
			name: 'MarketTaskAnalytics',
			meta: { title: 'Analytics', hidden: true },
			component: () => import('@/views/market/task/analytics.vue'),
		},
	],
}

export default route
