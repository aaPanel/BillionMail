<template>
	<div>
		<bt-table-layout>
			<template #toolsLeft>
				<n-button type="primary" @click="handleAdd">
					{{ t('common.actions.import') }}
				</n-button>
			</template>
			<template #toolsRight>
				<n-radio-group v-model:value="tableParams.status" @update:value="() => getTableData(true)">
					<n-radio-button :value="1">
						{{ t('contacts.subscribers.status.subscribed') }}
					</n-radio-button>
					<n-radio-button :value="0">
						{{ t('contacts.subscribers.status.unsubscribed') }}
					</n-radio-button>
				</n-radio-group>
				<div class="w-200px">
					<group-select
						v-model:value="tableParams.group_id"
						@update:value="() => getTableData(true)">
					</group-select>
				</div>
				<bt-search
					v-model:value="tableParams.keyword"
					:width="240"
					:placeholder="t('contacts.subscribers.search.emailPlaceholder')"
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
				<import-modal />
				<edit-modal />
			</template>
		</bt-table-layout>
	</div>
</template>

<script lang="tsx" setup>
import { useBrowserLocation } from '@vueuse/core'
import { DataTableColumns, NButton, NFlex } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { useTableData } from '@/hooks/useTableData'
import { confirm, formatTime } from '@/utils'
import { deleteSubscriber, getSubscriberList } from '@/api/modules/contacts/subscribers'
import type { Subscriber, SubscriberParams } from './interface'

import GroupSelect from './components/GroupSelect.vue'
import SubscriberImport from './components/SubscriberImport.vue'
import SubscriberEdit from './components/SubscriberEdit.vue'

const { t } = useI18n()

const location = useBrowserLocation()

const { tableParams, tableList, loading, tableTotal, getTableData } = useTableData<
	Subscriber,
	SubscriberParams
>({
	loading: true,
	immediate: true,
	params: {
		page: 1,
		page_size: 10,
		group_id: location.value.state.group_id || '',
		keyword: '',
		status: 1,
	},
	fetchFn: getSubscriberList,
})

// Table columns
const columns = ref<DataTableColumns<Subscriber>>([
	{
		key: 'email',
		title: t('contacts.subscribers.columns.email'),
		minWidth: 120,
		ellipsis: {
			tooltip: true,
		},
	},
	{
		key: 'groups',
		title: t('contacts.subscribers.columns.name'),
		minWidth: 100,
		render: row => {
			return row.groups.map(group => group.name).join(', ')
		},
	},
	{
		key: 'create_time',
		title: t('contacts.subscribers.columns.createdAt'),
		minWidth: 140,
		render: row => formatTime(row.create_time),
	},
	{
		title: t('common.columns.actions'),
		key: 'actions',
		align: 'right',
		width: 180,
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

const [ImportModal, importModalApi] = useModal({
	component: SubscriberImport,
	state: {
		refresh: getTableData,
	},
})

// Handle add subscriber
const handleAdd = () => {
	importModalApi.setState({ group_id: tableParams.value.group_id })
	importModalApi.open()
}

const [EditModal, editModalApi] = useModal({
	component: SubscriberEdit,
	state: {
		refresh: getTableData,
	},
})

// Handle edit
const handleEdit = (row: Subscriber) => {
	editModalApi.setState({ row })
	editModalApi.open()
}

// Handle delete
const handleDelete = (row: Subscriber) => {
	confirm({
		title: t('contacts.subscribers.delete.title'),
		content: t('contacts.subscribers.delete.confirm', { email: row.email }),
		confirmText: t('common.actions.delete'),
		confirmType: 'error',
		onConfirm: async () => {
			await deleteSubscriber({ emails: [row.email], status: tableParams.value.status })
			getTableData()
		},
	})
}
</script>

<style lang="scss" scoped>
.subscriber-container {
	padding: 16px;
	background-color: #fff;
	border-radius: 3px;

	.action-bar {
		display: flex;
		justify-content: space-between;
		margin-bottom: 16px;
	}
}
</style>
