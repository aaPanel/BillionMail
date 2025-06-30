import { instance } from '@/api'
import i18n from '@/i18n'

const { t } = i18n.global

export const getSystemConfig = () => {
	return instance.get('/settings/get_system_config')
}

export const getTimezoneList = () => {
	return instance.get('/settings/get_timezone_list')
}

export const setSystemConfigKey = (params: { key: string; value: string }) => {
	return instance.post('/settings/set_system_config_key', params, {
		fetchOptions: {
			loading: t('settings.api.loading.setting'),
			successMessage: true,
		},
	})
}

export const setSslConfig = (params: { certPem: string; privateKey: string }) => {
	return instance.post('/settings/set_ssl_config', params, {
		fetchOptions: {
			loading: t('settings.api.loading.settingSSL'),
			successMessage: true,
		},
	})
}

export const applyCert = () => {
	return instance.post(
		'/ssl/console_apply_cert',
		{},
		{
			fetchOptions: {
				loading: t('settings.api.loading.applyingCert'),
				successMessage: true,
			},
		}
	)
}

export const addIpWhitelist = (params: { ip: string }) => {
	return instance.post('/settings/add_ip_whitelist', params, {
		fetchOptions: {
			loading: t('settings.api.loading.adding'),
			successMessage: true,
		},
	})
}

export const deleteIpWhitelist = (params: { id: number }) => {
	return instance.post('/settings/delete_ip_whitelist', params, {
		fetchOptions: {
			loading: t('settings.api.loading.deleting'),
			successMessage: true,
		},
	})
}

export const clearIpWhitelist = () => {
	return instance.post(
		'/settings/set_ip_whitelist',
		{ ip_list: [] },
		{
			fetchOptions: {
				loading: t('settings.api.loading.clearing'),
				successMessage: true,
			},
		}
	)
}
