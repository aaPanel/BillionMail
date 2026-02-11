import { instance } from '@/api'

export interface VideoStatus {
	contact_email: string
	lead_tier: string
	lead_score: string
	video_url: string
	thumbnail_url: string
	landing_page_url: string
	status: string
}

export const generateVideo = (data: { contact_email: string; group_id: number }) => {
	return instance.post('/video_outreach/generate', data)
}

export const getVideoStatus = (params: { contact_email: string; group_id: number }) => {
	return instance.get('/video_outreach/status', { params })
}

export const listVideoOutreach = (params: {
	group_id?: number
	tier?: string
	page: number
	page_size: number
}) => {
	return instance.get('/video_outreach/list', { params })
}
