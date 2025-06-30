import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/contacts',
	redirect: '/contacts/group',
	name: 'ContactsLayout',
	component: Layout,
	meta: {
		sort: 4,
		key: 'contacts',
		title: 'Contacts',
		titleKey: 'layout.menu.contacts',
	},
	children: [
		{
			path: '/contacts',
			name: 'Contacts',
			redirect: '/contacts/group',
			component: () => import('@/views/contacts/index.vue'),
			children: [
				{
					path: 'group',
					name: 'ContactsGroup',
					meta: { title: 'Group', titleKey: 'layout.menu.group' },
					component: () => import('@/views/contacts/group/temp.vue'),
				},
				{
					path: 'subscribers',
					name: 'ContactsSubscribers',
					meta: { title: 'Subscribers', titleKey: 'layout.menu.subscribers' },
					component: () => import('@/views/contacts/subscribers/index.vue'),
				},
			],
		},
	],
}

export default route
