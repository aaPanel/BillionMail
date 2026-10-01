import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/video-outreach',
	name: 'VideoOutreachLayout',
	component: Layout,
	meta: {
		sort: 7,
		key: 'video-outreach',
		title: 'Video Outreach',
		titleKey: 'layout.menu.videoOutreach',
	},
	children: [
		{
			path: '/video-outreach',
			name: 'VideoOutreach',
			meta: { title: 'Video Outreach', titleKey: 'layout.menu.videoOutreach' },
			component: () => import('@/views/video-outreach/index.vue'),
		},
	],
}

export default route
