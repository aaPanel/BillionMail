<template>
	<div class="p-24px">
		<bt-table-layout>
			<template #toolsRight>
				<bt-search
					v-model:value="tableParams.keyword"
					:width="280"
					:placeholder="t('leads.search.placeholder')"
					@search="() => fetchTable(true)">
				</bt-search>
			</template>
			<template #table>
				<n-data-table v-bind="tableProps" :columns="columns" :row-key="(row: any) => row.id">
					<template #empty>
						<n-empty :description="t('leads.empty')" />
					</template>
				</n-data-table>
			</template>
			<template #pageRight>
				<bt-table-page v-bind="pageProps" @refresh="fetchTable"> </bt-table-page>
			</template>
		</bt-table-layout>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NTag } from 'naive-ui'
import { useDataTable } from '@/hooks/useDataTable'
import { getLeadList } from '@/api/modules/leads'

const { t } = useI18n()

const { tableProps, pageProps, tableParams, fetchTable } = useDataTable<any>({
	params: {
		page: 1,
		page_size: 10,
		keyword: '',
	},
	fetchFn: getLeadList,
})

const columns: DataTableColumns<any> = [
	{
		title: t('leads.columns.businessName'),
		key: 'business_name',
		ellipsis: { tooltip: true },
	},
	{
		title: t('leads.columns.ownerName'),
		key: 'owner_name',
		width: 150,
	},
	{
		title: t('leads.columns.ownerEmail'),
		key: 'owner_email',
		width: 200,
		ellipsis: { tooltip: true },
	},
	{
		title: t('leads.columns.status'),
		key: 'status',
		width: 120,
		render: (row) => {
			const typeMap: Record<string, string> = {
				scraped: 'default',
				enriched: 'info',
				personalized: 'success',
				queued: 'warning',
				in_sequence: 'primary',
				replied: 'success',
				booked: 'success',
				bounced: 'error',
			}
			return <NTag type={typeMap[row.status] || 'default'} size="small">{row.status}</NTag>
		},
	},
	{
		title: t('leads.columns.category'),
		key: 'category',
		width: 120,
	},
	{
		title: t('leads.columns.city'),
		key: 'city',
		width: 120,
	},
]
</script>
