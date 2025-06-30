<template>
	<div>
		<bt-table-layout>
			<template #toolsLeft>
				<n-button type="primary" @click="handleAdd">
					{{ t('market.task.actions.add') }}
				</n-button>
			</template>
			<template #toolsRight>
				<bt-search
					v-model:value="tableParams.keyword"
					:width="320"
					:placeholder="t('market.task.search.subjectPlaceholder')"
					@search="() => getTableData(true)">
				</bt-search>
			</template>
			<template #table>
				<n-data-table :loading="loading" :columns="columns" :data="tableList">
					<template #empty>
						<bt-table-help> </bt-table-help>
					</template>
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
			<template #modal>
				<status-modal />
				<detail-modal />
			</template>
		</bt-table-layout>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NButton, NFlex, NProgress, NTag } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { useTableData } from '@/hooks/useTableData'
import { confirm, formatDurationHighest, formatTime } from '@/utils'
import { deleteTask, getTaskList, pauseTask, resumeTask } from '@/api/modules/market/task'
import type { Task, TaskParams } from './interface'

import TaskDetail from './components/TaskDetail.vue'
import TaskStatus from './components/TaskStatus.vue'

const { t } = useI18n()

const { loading, tableParams, tableList, tableTotal, getTableData } = useTableData<
	Task,
	TaskParams
>({
	loading: true,
	immediate: true,
	params: {
		page: 1,
		page_size: 10,
		keyword: '',
	},
	fetchFn: getTaskList,
})

// Table columns
const columns = ref<DataTableColumns<Task>>([
	{
		key: 'create_time',
		title: t('market.task.columns.time'),
		width: '9%',
		minWidth: 130,
		ellipsis: {
			tooltip: true,
		},
		render: row => formatTime(row.create_time),
	},
	{
		key: 'subject',
		title: t('market.task.columns.subject'),
		minWidth: 120,
		ellipsis: {
			tooltip: true,
		},
	},
	{
		key: 'addresser',
		title: t('market.task.columns.sender'),
		minWidth: 130,
		ellipsis: {
			tooltip: true,
		},
	},
	{
		key: 'recipient_count',
		title: t('market.task.columns.recipients'),
		width: '6%',
		minWidth: 80,
	},
	{
		key: 'success_count',
		title: t('market.task.columns.success'),
		width: '6%',
		minWidth: 80,
	},
	{
		key: 'error_count',
		title: t('market.task.columns.failed'),
		width: '6%',
		minWidth: 80,
		render: row => (
			<NButton
				type="error"
				text
				onClick={() => {
					handleShowError(row)
				}}>
				{row.error_count}
			</NButton>
		),
	},
	{
		key: 'active',
		title: t('market.task.columns.status'),
		width: '8%',
		minWidth: 90,
		render: row => {
			if (row.task_process === 0 || row.task_process === 3)
				return (
					<NTag size="small" bordered={false} type="warning">
						{t('market.task.status.pending')}
					</NTag>
				)
			if (row.task_process === 1)
				return (
					<NTag size="small" bordered={false} type="info">
						{t('market.task.status.processing')}
					</NTag>
				)
			return (
				<NTag size="small" bordered={false} type="success">
					{t('market.task.status.done')}
				</NTag>
			)
		},
	},
	{
		key: 'remark',
		title: t('market.task.columns.remark'),
		width: '8%',
		minWidth: 80,
		ellipsis: {
			tooltip: true,
		},
		render: row => row.remark || '--',
	},
	{
		key: 'estimated_time_with_warmup',
		title: t('market.task.columns.estimatedTime'),
		width: '11%',
		minWidth: 110,
		ellipsis: {
			tooltip: true,
		},
		render: row => {
			// 已完成
			if (row.task_process === 2) {
				return '--'
			}
			// 显示预计时间
			if (row.estimated_time_with_warmup >= 0) {
				return formatDurationHighest(row.estimated_time_with_warmup)
			}
			return t('market.task.status.noIpWarmup')
		},
	},
	{
		key: 'progress',
		title: t('market.task.columns.progress'),
		width: '8%',
		minWidth: 130,
		render: row => {
			return (
				<div class="max-w-160px">
					<NProgress
						type="line"
						status="success"
						indicator-placement="inside"
						processing={row.progress !== 100}
						percentage={row.progress}
						show-indicator={false}
					/>
				</div>
			)
		},
	},
	{
		title: t('common.columns.actions'),
		key: 'actions',
		align: 'right',
		width: 220,
		render: row => (
			<NFlex inline={true}>
				{row.task_process !== 2 && (
					<NButton
						type="primary"
						text={true}
						onClick={() => {
							handleSetStatus(row)
						}}>
						{row.task_process === 3
							? t('market.task.actions.send')
							: t('market.task.actions.pause')}
					</NButton>
				)}
				<NButton
					type="primary"
					text={true}
					onClick={() => {
						handleGoAnalytics(row)
					}}>
					{t('market.task.actions.analytics')}
				</NButton>
				<NButton
					type="primary"
					text={true}
					onClick={() => {
						handleDetail(row)
					}}>
					{t('market.task.actions.detail')}
				</NButton>
				<NButton
					type="error"
					text={true}
					onClick={() => {
						handleDelete(row)
					}}>
					{t('common.actions.delete')}
				</NButton>
			</NFlex>
		),
	},
])

const router = useRouter()

// 添加任务
const handleAdd = () => {
	router.push('/market/task/edit')
}

// 前往分析
const handleGoAnalytics = (row: Task) => {
	router.push(`/market/task/analytics/${row.id}`)
}

// 暂停/发送任务
const handleSetStatus = (row: Task) => {
	confirm({
		title: t('market.task.changeStatus.title', { subject: row.subject }),
		content:
			row.task_process === 3
				? t('market.task.changeStatus.startConfirm')
				: t('market.task.changeStatus.pauseConfirm'),
		onConfirm: async () => {
			if (row.task_process === 3) {
				await resumeTask({ task_id: row.id })
			} else {
				await pauseTask({ task_id: row.id })
			}
			getTableData()
		},
	})
}

const [StatusModal, statusModalApi] = useModal({
	component: TaskStatus,
})

// 查看任务详情
const handleShowError = (row: Task) => {
	statusModalApi.setState({ row })
	statusModalApi.open()
}

const [DetailModal, detailModalApi] = useModal({
	component: TaskDetail,
})

// 查看任务详情
const handleDetail = (row: Task) => {
	detailModalApi.setState({ row })
	detailModalApi.open()
}

// 删除任务
const handleDelete = (row: Task) => {
	confirm({
		title: t('market.task.delete.title'),
		content: t('market.task.delete.confirm', { subject: row.subject }),
		confirmText: t('common.actions.delete'),
		confirmType: 'error',
		onConfirm: async () => {
			await deleteTask({ id: row.id })
			getTableData()
		},
	})
}
</script>

<style lang="scss" scoped></style>
