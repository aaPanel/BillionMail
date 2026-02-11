import { instance } from '@/api'

export const getSequenceList = (params: { page: number; page_size: number; keyword?: string }) => {
	return instance.get('/frostbyte/proxy/campaigns', { params })
}

export const getSequenceDetail = (id: string) => {
	return instance.get(`/frostbyte/proxy/campaigns/${id}`)
}
