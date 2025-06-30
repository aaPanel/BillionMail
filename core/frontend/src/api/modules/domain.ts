import { instance } from '@/api'
import i18n from '@/i18n'

const { t } = i18n.global

export const getDomainList = (params: { page: number; page_size: number; keyword: string }) => {
	return instance.get('/domains/list', { params })
}

export const getDomainAll = () => {
	return instance.get('/domains/all')
}

type DomainParams = {
	domain: string
	quota: number
	mailboxes: number
	email: string
}

export const createDomain = (params: DomainParams) => {
	return instance.post('/domains/create', params, {
		fetchOptions: {
			loading: t('domain.api.loading.creating'),
			successMessage: true,
		},
	})
}

export const updateDomain = (params: DomainParams) => {
	return instance.post('/domains/update', params, {
		fetchOptions: {
			loading: t('domain.api.loading.updating'),
			successMessage: true,
		},
	})
}

export const deleteDomain = (params: { domain: string }) => {
	return instance.post('/domains/delete', params, {
		fetchOptions: {
			loading: t('domain.api.loading.deleting'),
			successMessage: true,
		},
	})
}

export const setSsl = (params: { domain: string; certificate: string; key: string }) => {
	return instance.post('/domains/set_ssl', params, {
		fetchOptions: {
			loading: t('domain.api.loading.settingSSL'),
			successMessage: true,
		},
	})
}

export const getSsl = (params: { domain: string }) => {
	return instance.get('/domains/get_ssl', { params })
}

export const freshDnsRecord = (params: { domain: string }) => {
	return instance.post('/domains/fresh_dns_records', params, {
		fetchOptions: {
			loading: t('domain.api.loading.refreshingDns'),
			successMessage: true,
		},
	})
}

export const applyCert = (params: { domain: string }) => {
	return instance.post('/ssl/apply_cert', params, {
		fetchOptions: {
			loading: t('domain.api.loading.applyingCert'),
			successMessage: true,
		},
	})
}

export const setDefaultDomain = (params: { domain: string }) => {
	return instance.post('/domains/set_default_domain', params, {
		fetchOptions: {
			loading: 'Setting Default Domain, please wait...',
			successMessage: true,
		},
	})
}
