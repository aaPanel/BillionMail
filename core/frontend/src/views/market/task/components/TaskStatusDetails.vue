<template>
	<modal :title="title" :width="1000" :footer="false">
		<bt-table-layout>
			<template #table>
				<n-data-table
					:bordered="false"
					:loading="loading"
					:columns="columns"
					:data="tableList"
					:row-props="rowProps">
				</n-data-table>
			</template>
			<template #pageRight>
				<bt-table-page
					v-model:page="tableParams.page"
					v-model:page-size="tableParams.page_size"
					:item-count="tableTotal"
					@refresh="getTableData">
				</bt-table-page>
			</template>
		</bt-table-layout>
	</modal>
</template>

<script lang="tsx" setup>
import { DataTableColumns, DataTableCreateRowProps, NEllipsis } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { useTableData } from '@/hooks/useTableData'
import { getMailProviderLogs } from '@/api/modules/market/task'
import type { TaskStatus, TaskStatusDetails } from '../interface'

interface TaskStatusDetailsParams {
	page: number
	page_size: number
	task_id: number
	status: number
	domain: string
}

const title = ref('')

const { loading, tableParams, tableList, tableTotal, getTableData } = useTableData<
	TaskStatusDetails,
	TaskStatusDetailsParams
>({
	loading: true,
	immediate: false,
	params: {
		page: 1,
		page_size: 10,
		task_id: 0,
		status: 0,
		domain: '',
	},
	fetchFn: getMailProviderLogs,
})

const columns = ref<DataTableColumns<TaskStatusDetails>>([
	{
		key: 'recipient',
		title: 'Recipient',
		width: 160,
		render: row => {
			return (
				<NEllipsis
					class={row.tooltip ? '' : 'break-all!'}
					line-clamp={row.tooltip ? 999 : 1}
					tooltip={false}>
					{row.recipient}
				</NEllipsis>
			)
		},
	},
	{
		key: 'dsn',
		title: 'Dsn',
		width: 60,
	},
	{
		key: 'delay',
		title: 'Delay',
		width: 70,
	},
	{
		key: 'delays',
		title: 'Delays',
		width: 150,
	},
	{
		key: 'relay',
		title: 'Relay',
		width: 200,
		render: row => {
			return (
				<NEllipsis
					class={row.tooltip ? '' : 'break-all!'}
					line-clamp={row.tooltip ? 9999 : 1}
					tooltip={false}>
					{row.relay}
				</NEllipsis>
			)
		},
	},
	{
		key: 'description',
		title: 'Details',
		render: row => {
			return (
				<NEllipsis
					class={row.tooltip ? '' : 'break-all!'}
					line-clamp={row.tooltip ? 9999 : 1}
					tooltip={false}>
					{row.description}
				</NEllipsis>
			)
		},
	},
])

const rowProps: DataTableCreateRowProps<TaskStatusDetails> = row => {
	return {
		onMousemove: () => {
			row.tooltip = true
		},
		onMouseleave: () => {
			row.tooltip = false
		},
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ status: number; taskId: number; row: TaskStatus }>()
			const { status, taskId, row } = state
			if (row) {
				title.value = `Failure Details [${row.domain}]`
				tableParams.value.page = 1
				tableParams.value.status = status
				tableParams.value.task_id = taskId
				tableParams.value.domain = row.domain
				getTableData()
			}
		} else {
			tableList.value = []
		}
	},
})
</script>

<style lang="scss" scoped></style>
