<template>
	<bt-table-layout>
		<template #toolsLeft>
			<n-button type="primary" disabled>
				{{ t('common.actions.add') }}
			</n-button>
		</template>
		<template #toolsRight>
			<bt-search
				v-model:value="tableParams.keyword"
				:width="280"
				:placeholder="t('sequences.search.placeholder')"
				@search="() => fetchTable(true)">
			</bt-search>
		</template>
		<template #table>
			<n-data-table v-bind="tableProps" :columns="columns">
				<template #empty>
					<n-empty :description="t('sequences.empty')" />
				</template>
			</n-data-table>
		</template>
		<template #pageRight>
			<bt-table-page v-bind="pageProps" @refresh="fetchTable"> </bt-table-page>
		</template>
	</bt-table-layout>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NTag } from 'naive-ui'
import { useDataTable } from '@/hooks/useDataTable'
import { getSequenceList } from '@/api/modules/sequences'
import { formatTime } from '@/utils'

const { t } = useI18n()
const router = useRouter()

const { tableProps, pageProps, tableParams, fetchTable } = useDataTable<any>({
	params: {
		page: 1,
		page_size: 10,
		keyword: '',
	},
	fetchFn: getSequenceList,
})

const columns: DataTableColumns<any> = [
	{
		title: t('sequences.columns.name'),
		key: 'name',
		ellipsis: { tooltip: true },
	},
	{
		title: t('sequences.columns.status'),
		key: 'status',
		width: 120,
		render: (row) => {
			const typeMap: Record<string, string> = {
				active: 'success',
				paused: 'warning',
				draft: 'default',
				completed: 'info',
			}
			return <NTag type={typeMap[row.status] || 'default'} size="small">{row.status}</NTag>
		},
	},
	{
		title: t('sequences.columns.steps'),
		key: 'step_count',
		width: 80,
	},
	{
		title: t('sequences.columns.sent'),
		key: 'total_sent',
		width: 80,
	},
	{
		title: t('sequences.columns.opens'),
		key: 'total_opens',
		width: 80,
	},
	{
		title: t('sequences.columns.replies'),
		key: 'total_replies',
		width: 80,
	},
	{
		title: t('sequences.columns.created'),
		key: 'created_at',
		width: 160,
		render: (row) => formatTime(row.created_at),
	},
	{
		title: t('common.columns.actions'),
		key: 'actions',
		width: 100,
		render: (row) => (
			<n-button
				text
				type="primary"
				onClick={() => router.push(`/sequences/${row.id}`)}>
				{t('common.actions.view')}
			</n-button>
		),
	},
]
</script>
