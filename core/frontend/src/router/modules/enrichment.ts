import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/enrichment',
	name: 'EnrichmentLayout',
	component: Layout,
	meta: {
		sort: 7,
		key: 'enrichment',
		title: 'Enrichment',
		titleKey: 'layout.menu.enrichment',
	},
	children: [
		{
			path: '/enrichment',
			name: 'Enrichment',
			meta: { title: 'Enrichment', titleKey: 'layout.menu.enrichment' },
			component: () => import('@/views/enrichment/index.vue'),
		},
	],
}

export default route
