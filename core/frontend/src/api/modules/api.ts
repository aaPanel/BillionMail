import { instance } from '@/api'
import i18n from '@/i18n'

const { t } = i18n.global

export const getApiList = (params: {
	page: number
	page_size: number
	keyword: string
	active: number
	start_time: number
	end_time: number
}) => {
	return instance.get('/batch_mail/api/list', { params })
}

export const getOverviewStats = (params: { start_time: number; end_time: number }) => {
	return instance.get('/batch_mail/api/overview_stats', { params })
}

type APIParams = {
	api_name: string
	template_id: number
	subject: string
	addresser: string
	full_name: string
	unsubscribe: number
	active: number
	ip_whitelist: string[]
}

export const createApi = (params: APIParams) => {
	return instance.post('/batch_mail/api/create', params, {
		fetchOptions: {
			loading: t('api.loading.creating'),
			successMessage: true,
		},
	})
}

export const updateApi = (params: APIParams & { id: number }) => {
	return instance.post('/batch_mail/api/update', params, {
		fetchOptions: {
			loading: t('api.loading.updating'),
			successMessage: true,
		},
	})
}

export const deleteApi = (params: { id: number }) => {
	return instance.post('/batch_mail/api/delete', params, {
		fetchOptions: {
			loading: t('api.loading.deleting'),
			successMessage: true,
		},
	})
}

export const testApi = (key: string, params: { recipient: string }) => {
	return instance.post('/batch_mail/api/send', params, {
		headers: {
			'x-api-key': key,
		},
		fetchOptions: {
			loading: t('api.loading.testing'),
			successMessage: true,
		},
	})
}
