<template>
	<div>
		<bt-table-layout>
			<template #toolsLeft>
				<n-button type="primary" @click="handleAdd">
					{{ t('common.actions.add') }}
				</n-button>
			</template>
			<template #toolsRight>
				<bt-search
					v-model:value="tableParams.keyword"
					:width="280"
					:placeholder="t('contacts.group.search.namePlaceholder')"
					@search="() => getTableData(true)">
				</bt-search>
			</template>
			<template #table>
				<n-data-table
					v-model:checked-row-keys="checkedKeys"
					:row-key="row => row.id"
					:loading="loading"
					:columns="columns"
					:data="tableList">
					<template #empty>
						<bt-table-help> </bt-table-help>
					</template>
				</n-data-table>
			</template>
			<template #pageLeft>
				<n-button :disabled="checkedKeys.length === 0" @click="handleExport">
					{{ $t('common.actions.export') }}
				</n-button>
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
				<add-modal />
				<rename-modal />
			</template>
		</bt-table-layout>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NButton, NFlex } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { useTableData } from '@/hooks/useTableData'
import { confirm, formatTime, isObject } from '@/utils'
import { getGroupList, deleteGroup, exportGroup } from '@/api/modules/contacts/group'
import type { Group, GroupParams } from './types/base.ts'

import GroupAdd from './components/GroupAdd.vue'
import GroupRename from './components/GroupRename.vue'
import { downloadFile } from '@/api/modules/public'

const { t } = useI18n()

const router = useRouter()

const checkedKeys = ref<number[]>([])

const { tableParams, tableList, loading, tableTotal, getTableData } = useTableData<
	Group,
	GroupParams
>({
	loading: true,
	immediate: true,
	params: {
		page: 1,
		page_size: 10,
		keyword: '',
	},
	fetchFn: getGroupList,
})

// Table columns
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
		render: row => row.total_count || 0,
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
					{t('common.actions.rename')}
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

const [AddModal, addModalApi] = useModal({
	component: GroupAdd,
	state: {
		refresh: getTableData,
	},
})

// Handle add group
const handleAdd = () => {
	addModalApi.open()
}

const [RenameModal, renameModalApi] = useModal({
	component: GroupRename,
	state: {
		refresh: getTableData,
	},
})

// Handle edit
const handleEdit = (row: Group) => {
	renameModalApi.setState({ row })
	renameModalApi.open()
}

// Handle delete
const handleDelete = (row: Group) => {
	confirm({
		title: t('contacts.group.delete.title'),
		content: t('contacts.group.delete.confirm', { name: row.name }),
		confirmText: t('common.actions.delete'),
		confirmType: 'error',
		onConfirm: async () => {
			await deleteGroup({ group_ids: [row.id] })
			getTableData()
		},
	})
}

const handleExport = () => {
	confirm({
		title: 'Export group',
		content: `Export ${checkedKeys.value.length} group(s)?`,
		onConfirm: async () => {
			const res = await exportGroup({
				format: 'csv',
				include_unsubscribe: true,
				group_ids: checkedKeys.value,
				export_type: 1,
			})
			if (isObject<{ file_url: string }>(res)) {
				downloadFile({ file_path: res.file_url })
			}
		},
	})
}
</script>
