import i18n from '@/i18n'
import { instance } from '@/api'
import type { TaskParams } from '@/views/market/task/interface'

const { t } = i18n.global

/**
 * 获取任务列表
 */
export function getTaskList(params: TaskParams) {
	return instance.get('/batch_mail/task/list', { params })
}

/**
 * 获取任务统计
 */
export function getTaskOverview(params: { task_id: number; start_time: number; end_time: number }) {
	return instance.get('/batch_mail/task/stat_chart', { params })
}

/**
 * 获取任务详情
 */
export function getTaskDetails(params: { id: number }) {
	return instance.get('/batch_mail/task/find', { params })
}

interface TaskAddParams {
	track_open: number
	track_click: number
	addresser: string
	full_name: string
	subject: string
	group_ids: number[]
	template_id: number
	is_record: number
	warmup: number
	unsubscribe: number
	threads: number
	start_time: number
	remark: string
}

/**
 * 获取任务列表
 */
export function addTask(params: TaskAddParams) {
	return instance.post('/batch_mail/task/create', params, {
		fetchOptions: {
			loading: t('market.task.loading.creating'),
			successMessage: true,
		},
	})
}

/**
 * 删除任务
 */
export function deleteTask(params: { id: number }) {
	return instance.post('/batch_mail/task/delete', params, {
		fetchOptions: {
			loading: t('market.task.loading.deleting'),
			successMessage: true,
		},
	})
}

/**
 * 暂停任务
 */
export function pauseTask(params: { task_id: number }) {
	return instance.post('/batch_mail/task/pause', params, {
		fetchOptions: {
			loading: t('market.task.loading.pausing'),
			successMessage: true,
		},
	})
}

/**
 * 恢复任务
 */
export function resumeTask(params: { task_id: number }) {
	return instance.post('/batch_mail/task/resume', params, {
		fetchOptions: {
			loading: t('market.task.loading.resuming'),
			successMessage: true,
		},
	})
}

/**
 * 恢复任务
 */
export function sendTestEmail(params: {
	addresser: string
	subject: string
	recipient: string
	template_id: number
}) {
	return instance.post('/batch_mail/task/send_test', params, {
		fetchOptions: {
			loading: t('market.task.loading.sendingTest'),
			successMessage: true,
		},
	})
}

export function getMailProvider(params: { task_id: number; status: number }) {
	return instance.get('/batch_mail/tracking/mail_provider', { params })
}

export function getMailProviderLogs(params: {
	page: number
	page_size: number
	task_id: number
	domain: string
}) {
	return instance.get('/batch_mail/tracking/logs', { params })
}
