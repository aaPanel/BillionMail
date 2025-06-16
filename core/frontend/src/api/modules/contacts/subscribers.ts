import { instance } from '@/api'
import { SubscriberParams } from '@/views/contacts/subscribers/interface'
import i18n from '@/i18n'

const { t } = i18n.global

export const getSubscriberList = (params: SubscriberParams) => {
	return instance.get('/contact/list', { params })
}

export const importSubscribers = (data: {
	group_ids: number[]
	file_data: string
	file_type: string
	contacts: string
	import_type: number
	overwrite: number
	default_active: number
	status: number
}) => {
	return instance.post('/contact/group/import', data, {
		fetchOptions: {
			loading: t('contacts.group.loading.importSubscribers'),
			successMessage: true,
		},
	})
}

export const updateSubscriberGroup = (data: {
	emails: string[]
	active: number
	attribs: string
	group_ids: number[]
}) => {
	return instance.post('/contact/update_group', data, {
		fetchOptions: {
			loading: t('contacts.group.loading.updateSubscriberGroup'),
			successMessage: true,
		},
	})
}

export const editContact = (data: {
	emails: string
	active: number
	attribs: string
	group_ids: number[]
}) => {
	return instance.post('/contact/edit', data, {
		fetchOptions: {
			loading: t('contacts.group.loading.updateSubscriberGroup'),
			successMessage: true,
		},
	})
}

export const deleteSubscriber = (data: { emails: string[]; status: number }) => {
	return instance.post('/contact/delete', data, {
		fetchOptions: {
			loading: t('contacts.group.loading.deleteSubscriber'),
			successMessage: true,
		},
	})
}
