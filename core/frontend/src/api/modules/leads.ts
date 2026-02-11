import { instance } from '@/api'

export const getLeadList = (params: { page: number; page_size: number; keyword?: string }) => {
	return instance.get('/frostbyte/proxy/leads', { params })
}

export const getLeadDetail = (id: number) => {
	return instance.get(`/frostbyte/proxy/leads/${id}`)
}
