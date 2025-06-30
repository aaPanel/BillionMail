import i18n from '@/i18n'
import { instance } from '@/api'
import type { TemplateParams } from '@/views/market/template/interface'

const { t } = i18n.global

export function getTemplateList(params: TemplateParams) {
	return instance.get('/email_template/list', { params })
}

export function getTemplateAll() {
	return instance.get('/email_template/all')
}

export function getTemplateDetails(params: { id: string }) {
	return instance.get('/email_template/get', { params })
}

interface TemplateForm {
	temp_name: string
	add_type: number
	html_content: string
	drag_data: string
}

export function addTemplate(params: TemplateForm) {
	return instance.post('/email_template/create', params, {
		fetchOptions: {
			loading: t('market.template.loading.creating'),
			successMessage: true,
		},
	})
}

export function updateTemplate(params: TemplateForm & { id: number }) {
	return instance.post('/email_template/update', params, {
		fetchOptions: {
			loading: t('market.template.loading.updating'),
			successMessage: true,
		},
	})
}

export function duplicateTemplate(params: { id: number }) {
	return instance.post('/email_template/copy', params, {
		fetchOptions: {
			loading: t('market.template.loading.copying'),
			successMessage: true,
		},
	})
}

export function deleteTemplate(params: { id: number }) {
	return instance.post('/email_template/delete', params, {
		fetchOptions: {
			loading: t('market.template.loading.deleting'),
			successMessage: true,
		},
	})
}
