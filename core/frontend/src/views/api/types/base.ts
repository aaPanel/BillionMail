export interface Api {
	id: number
	api_key: string
	api_name: string
	template_id: number
	subject: string
	addresser: string
	full_name: string
	unsubscribe: number
	track_open: number
	track_click: number
	active: number
	create_time: number
	update_time: number
	expire_time: number
	send_count: number
	success_count: number
	fail_count: number
	open_rate: number
	click_rate: number
	delivery_rate: number
	bounce_rate: number
	ip_whitelist: string[]
}

export interface OverviewStats {
	total_send: number
	avg_delivery_rate: number
	avg_open_rate: number
	avg_click_rate: number
	avg_bounce_rate: number
	avg_unsub_rate: number
	total_unsubscribe: number
}
