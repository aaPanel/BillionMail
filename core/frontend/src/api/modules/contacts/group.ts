import { instance } from '@/api'
import type { GroupParams } from '@/views/contacts/group/interface'
import i18n from '@/i18n'

const { t } = i18n.global

export const getGroupList = (params: GroupParams) => {
	return instance.get('/contact/group/list', { params })
}

export const getGroupAll = () => {
	return instance.get('/contact/group/all')
}

export const createGroup = (data: {
	create_type: number
	name: string
	description: string
	status: number
	file_data?: string
	file_type?: string
}) => {
	return instance.post('/contact/group/create', data, {
		fetchOptions: {
			loading: t('contacts.group.loading.createGroup'),
			successMessage: true,
		},
	})
}

export const updateGroup = (data: { group_id: number; name: string }) => {
	return instance.post('/contact/group/update', data, {
		fetchOptions: {
			loading: t('contacts.group.loading.updateGroup'),
			successMessage: true,
		},
	})
}

export const deleteGroup = (data: { group_ids: number[] }) => {
	return instance.post('/contact/group/delete', data, {
		fetchOptions: {
			loading: t('contacts.group.loading.deleteGroup'),
			successMessage: true,
		},
	})
}

export const getContactCount = (data: { group_ids: number[] }) => {
	return instance.post('/contact/group/contact_count', data)
}

export const exportGroup = (data: {
	format: string
	include_unsubscribe: boolean
	group_ids: number[]
	export_type: number
}) => {
	return instance.post('/contact/group/export', data, {
		fetchOptions: {
			loading: 'Exporting, please wait...',
			successMessage: true,
		},
	})
}
