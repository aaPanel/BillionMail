import { instance } from '@/api'
import type { TemplateParams } from '@/views/market/template/interface'

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
			loading: 'Creating template, please wait...',
			successMessage: true,
		},
	})
}

export function updateTemplate(params: TemplateForm & { id: number }) {
	return instance.post('/email_template/update', params, {
		fetchOptions: {
			loading: 'Updating template, please wait...',
			successMessage: true,
		},
	})
}

export function duplicateTemplate(params: { id: number }) {
	return instance.post('/email_template/copy', params, {
		fetchOptions: {
			loading: 'Copying template, please wait...',
			successMessage: true,
		},
	})
}

export function deleteTemplate(params: { id: number }) {
	return instance.post('/email_template/delete', params, {
		fetchOptions: {
			loading: 'Deleting template, please wait...',
			successMessage: true,
		},
	})
}
