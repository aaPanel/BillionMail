export interface Task {
	id: number
	task_name: string
	addresser: string
	subject: string
	full_name: string
	recipient_count: number
	task_process: number
	pause: number
	template_id: number
	is_record: number
	unsubscribe: number
	threads: number
	etypes: string
	track_open: number
	track_click: number
	start_time: number
	create_time: number
	update_time: number
	remark: string
	active: number
	groups: Group[]
	progress: number
	sent_count: number
	unsent_count: number
	template_name: string
	success_count: number
	error_count: number
	estimated_time_with_warmup: number
}

export interface Group {
	id: number
	name: string
	description: string
	count: number
}

export interface TaskParams {
	page: number
	page_size: number
	keyword: string
}

export interface TaskDetailItem {
	id: number
	email: string
	status: number
	error_msg: string
	create_time: number
}

export interface TaskDetailParams {
	task_id: number
	page: number
	page_size: number
	status?: number
}

export interface TaskStatus {
	count: number
	domain: string
}

export interface TaskStatusDetails {
	postfix_message_id: string
	status: string
	recipient: string
	mail_provider: string
	delay: number
	delays: string
	dsn: string
	relay: string
	description: string
	log_time: number
	tooltip: boolean
}
