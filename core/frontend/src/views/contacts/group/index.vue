<template>
	<bt-table-layout>
		<template #toolsLeft>
			<n-button type="primary" @click="handleAdd">Add group</n-button>
		</template>
		<template #table>
			<n-data-table v-bind="tableProps" :columns="columns"></n-data-table>
		</template>
		<template #pageRight>
			<bt-table-page v-bind="pageProps" @refresh="fetchTable"> </bt-table-page>
		</template>
	</bt-table-layout>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NButton, NFlex, NTag } from 'naive-ui'
import { formatTime } from '@/utils'
import { getGroupList } from '@/api/modules/contacts/group'
import { useDataTable } from './hooks/useDataTable'
import type { Group, GroupParams } from './types/base.ts'

const { t } = useI18n()

const router = useRouter()

const { tableProps, pageProps, fetchTable } = useDataTable<Group, GroupParams>({
	params: {
		page: 1,
		page_size: 10,
		keyword: '',
	},
	fetchFn: getGroupList,
})

const columns = ref<DataTableColumns<Group>>([
	{
		type: 'selection',
	},
	{
		key: 'name',
		title: t('contacts.group.columns.name'),
		minWidth: 120,
		render: row => (
			<NButton
				text
				type="primary"
				onClick={() => {
					router.push({
						path: '/contacts/subscribers',
						state: { group_id: row.id },
					})
				}}>
				{row.name}
			</NButton>
		),
	},
	{
		key: 'total_count',
		title: t('contacts.group.columns.subscribers'),
		minWidth: 100,
		render: row => (
			<NTag size="small" type="primary" bordered={false}>
				{row.total_count || 0}
			</NTag>
		),
	},
	{
		key: 'unsubscribe_count',
		title: t('contacts.group.columns.unsubscribe'),
		minWidth: 100,
		render: row => row.unsubscribe_count || 0,
	},
	{
		key: 'create_time',
		title: t('contacts.group.columns.createdAt'),
		minWidth: 140,
		render: row => formatTime(row.create_time),
	},
	{
		title: t('common.columns.actions'),
		key: 'actions',
		align: 'right',
		width: 120,
		render: row => (
			<NFlex inline={true}>
				<NButton
					type="primary"
					text={true}
					onClick={() => {
						handleEdit(row)
					}}>
					{t('common.actions.edit')}
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

const handleAdd = () => {
	console.log(1)
}

const handleEdit = (row: Group) => {
	console.log(row)
}

const handleDelete = (row: Group) => {
	console.log(row)
}
</script>

<style lang="scss" scoped></style>
