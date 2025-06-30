<template>
	<bt-table-layout>
		<template #toolsLeft>
			<n-button type="primary" @click="handleAdd">{{ $t('settings.forward.add') }}</n-button>
		</template>
		<template #toolsRight>
			<n-button @click="handleRefresh">
				{{ $t('common.actions.refresh') }}
			</n-button>
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
			<form-modal />
		</template>
	</bt-table-layout>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NButton, NFlex, NSwitch } from 'naive-ui'
import { confirm, formatTime } from '@/utils'
import { useTableData } from '@/hooks/useTableData'
import { useModal } from '@/hooks/modal/useModal'
import { deleteForward, editForward, getForwardList } from '@/api/modules/settings/forward'
import type { Forward, ForwardParams } from './types/base'

import ForwardForm from './components/ForwardForm.vue'

const { t } = useI18n()

const { tableParams, tableList, loading, tableTotal, getTableData } = useTableData<
	Forward,
	ForwardParams
>({
	loading: true,
	immediate: true,
	params: {
		page: 1,
		page_size: 10,
		domain: '',
		keyword: '',
	},
	fetchFn: getForwardList,
})

const columns = ref<DataTableColumns<Forward>>([
	{
		key: 'address',
		title: t('settings.forward.columns.address'),
		ellipsis: {
			tooltip: true,
		},
	},
	{
		key: 'goto',
		title: t('settings.forward.columns.goto'),
		ellipsis: {
			tooltip: true,
		},
	},
	// {
	// 	key: 'domain',
	// 	title: t('settings.forward.columns.domain'),
	// 	ellipsis: {
	// 		tooltip: true,
	// 	},
	// },
	{
		key: 'create_time',
		title: t('settings.forward.columns.created'),
		render: row => {
			return formatTime(row.create_time)
		},
	},
	{
		key: 'update_time',
		title: t('settings.forward.columns.modified'),
		render: row => {
			return formatTime(row.update_time)
		},
	},
	{
		key: 'active',
		title: t('settings.forward.columns.status'),
		render: row => (
			<NSwitch
				size="small"
				value={row.active}
				checked-value={1}
				unchecked-value={0}
				onUpdateValue={val => {
					handleSetStatus(val, row)
				}}
			/>
		),
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

const [FormModal, formModalApi] = useModal({
	component: ForwardForm,
	state: {
		isEdit: false,
		refresh: getTableData,
	},
})

const handleAdd = () => {
	formModalApi.setState({ isEdit: false, row: null })
	formModalApi.open()
}

const handleRefresh = () => {
	getTableData(true)
}

const handleEdit = (row: Forward) => {
	formModalApi.setState({ isEdit: true, row })
	formModalApi.open()
}

const handleSetStatus = async (val: number, row: Forward) => {
	await editForward({ address: row.address, goto: row.goto, active: val })
	getTableData()
}

const handleDelete = (row: Forward) => {
	confirm({
		title: t('settings.forward.delete.title'),
		content: t('settings.forward.delete.content', { address: row.address }),
		confirmText: t('common.actions.delete'),
		confirmType: 'error',
		onConfirm: async () => {
			await deleteForward({ address: row.address })
			getTableData()
		},
	})
}
</script>

<style lang="scss" scoped></style>
