import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/sequences',
	redirect: '/sequences/list',
	name: 'SequencesLayout',
	component: Layout,
	meta: {
		sort: 5,
		key: 'sequences',
		title: 'Sequences',
		titleKey: 'layout.menu.sequences',
	},
	children: [
		{
			path: '/sequences',
			name: 'Sequences',
			redirect: '/sequences/list',
			component: () => import('@/views/sequences/index.vue'),
			children: [
				{
					path: 'list',
					name: 'SequencesList',
					meta: { title: 'Sequences', titleKey: 'layout.menu.sequences' },
					component: () => import('@/views/sequences/list/index.vue'),
				},
			],
		},
		{
			path: ':id',
			name: 'SequenceDetail',
			meta: { title: 'Sequence Detail', titleKey: '' },
			component: () => import('@/views/sequences/detail/index.vue'),
		},
	],
}

export default route
