export interface Group {
	id: number
	name: string
	description: string
	create_time: number
	update_time: number
	token: string
	double_optin: number
	welcome_mail_id: number
	confirm_mail_id: number
	success_url: string
	confirm_url: string
	already_url: string
	total_count: number
	active_count: number
	unsubscribe_count: number
	subscribe_link: string
	subscribe_form: string
}

export interface GroupParams {
	page: number
	page_size: number
	keyword: string
}
