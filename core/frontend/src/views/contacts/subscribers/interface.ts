export interface Subscriber {
	id: number
	email: string
	group_id: number
	active: number
	task_id: number
	create_time: number
	attribs: Record<string, string> | null
	groups: Array<{
		id: number
		name: string
	}>
}

export interface SubscriberParams {
	page: number
	page_size: number
	group_id: string
	keyword: string
	status: number
}
