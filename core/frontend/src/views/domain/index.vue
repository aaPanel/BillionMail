<template>
	<div class="p-24px">
		<div class="bt-title">{{ t('layout.menu.domain') }}</div>
		<bt-table-layout>
			<template #toolsLeft>
				<n-button type="primary" @click="handleAddDomain">{{ t('domain.addDomain') }}</n-button>
				<bt-help href="https://www.billionmail.com/start/domain.html" :text="t('domain.help')">
				</bt-help>
			</template>
			<template #table>
				<n-data-table :loading="loading" :columns="columns" :data="tableList">
					<template #empty>
						<bt-table-help></bt-table-help>
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
				<ssl-modal />
				<dns-modal />
			</template>
		</bt-table-layout>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NFlex, NButton, NTag } from 'naive-ui'
import { confirm, getByteUnit } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { useTableData } from '@/hooks/useTableData'
import { deleteDomain, getDomainList, setDefaultDomain } from '@/api/modules/domain'
import type { MailDomain, MailDomainParams } from './interface'

import DomainForm from './components/DomainForm.vue'
import DomainSsl from './components/DomainSsl/index.vue'
import DomainDns from './components/DomainDns.vue'

const { t } = useI18n()

const { tableParams, tableList, loading, tableTotal, getTableData } = useTableData<
	MailDomain,
	MailDomainParams
>({
	immediate: true,
	params: {
		page: 1,
		page_size: 10,
		keyword: '',
	},
	fetchFn: getDomainList,
})

const router = useRouter()

// Table columns
const columns = ref<DataTableColumns<MailDomain>>([
	{
		key: 'domain',
		title: t('domain.columns.domain'),
		minWidth: 130,
		ellipsis: {
			tooltip: true,
		},
		render: row => {
			return (
				<>
					<NButton
						text
						type="primary"
						onClick={() => {
							router.push({
								path: '/mailbox',
								state: { domain: row.domain },
							})
						}}>
						{row.domain}
					</NButton>
					{row.default ? (
						<NTag size="small" class="ml-8px" bordered={false}>
							Default
						</NTag>
					) : (
						''
					)}
				</>
			)
		},
	},
	{
		key: 'quota',
		title: t('domain.columns.quota'),
		render: row => getByteUnit(row.quota),
	},
	{
		key: 'mailboxes',
		title: t('domain.columns.mailboxes'),
	},
	{
		key: 'mailbox_quota',
		title: t('domain.columns.defaultMailboxSize'),
		render: row => getByteUnit(row.mailbox_quota),
	},
	// {
	// 	key: 'catch_all',
	// 	title: 'Catch All',
	// 	render: row => (
	// 		<NSwitch
	// 			value={row.catch_all}
	// 			size="small"
	// 			onUpdateValue={(val: boolean) => {
	// 				if (val) {
	// 					handleOpenCatch(row)
	// 				} else {
	// 					handleCloseCatch(row)
	// 				}
	// 			}}
	// 		/>
	// 	),
	// },
	{
		key: 'ssl',
		title: 'SSL',
		render: row => {
			if (row.cert_info && row.cert_info.endtime) {
				const ssl = row.cert_info
				const day = Math.floor((ssl.endtime * 1000 - Date.now()) / 1000 / 60 / 60 / 24)
				return (
					<NButton
						type={day < 0 ? 'error' : 'primary'}
						text
						onClick={() => {
							handleShowSsl(row)
						}}>
						{day < 0 ? t('domain.ssl.expired') : t('domain.ssl.daysLeft', { days: day })}
					</NButton>
				)
			}
			return (
				<NButton
					type="warning"
					text
					onClick={() => {
						handleShowSsl(row)
					}}>
					{t('domain.ssl.notSet')}
				</NButton>
			)
		},
	},
	{
		title: t('domain.columns.actions'),
		key: 'actions',
		align: 'right',
		width: 260,
		render: row => (
			<NFlex inline={true}>
				<NButton
					type="primary"
					text={true}
					onClick={() => {
						handleDNSRecord(row)
					}}>
					{t('domain.actions.dnsRecord')}
				</NButton>
				<NButton
					type="primary"
					text={true}
					onClick={() => {
						handleEdit(row)
					}}>
					{t('common.actions.edit')}
				</NButton>
				<NButton
					type="primary"
					text={true}
					disabled={row.default === 1}
					onClick={() => {
						handleSetDefault(row)
					}}>
					Set Default
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
	component: DomainForm,
	state: {
		isEdit: false,
		refresh: getTableData,
	},
})

// Handle add domain
const handleAddDomain = () => {
	formModalApi.setState({
		row: null,
		isEdit: false,
	})
	formModalApi.open()
}

// Handle open catch all
// const handleOpenCatch = (row: MailDomain) => {
// 	catchModalApi.setState({ row })
// 	catchModalApi.open()
// }

// Handle close catch all
// const handleCloseCatch = (row: MailDomain) => {
// 	console.log(row)
// }

const [SslModal, sslModalApi] = useModal({
	component: DomainSsl,
	state: {
		refresh: getTableData,
	},
})

// Handle show ssl
const handleShowSsl = (row: MailDomain) => {
	sslModalApi.setState({ row })
	sslModalApi.open()
}

const [DnsModal, dnsModalApi] = useModal({
	component: DomainDns,
	state: {
		refresh: getTableData,
	},
})

// Handle DNS records
const handleDNSRecord = (row: MailDomain) => {
	dnsModalApi.setState({ row })
	dnsModalApi.open()
}

const handleSetDefault = (row: MailDomain) => {
	confirm({
		title: `Set Default [${row.domain}]`,
		content: 'Are you sure to set this domain as default?',
		onConfirm: async () => {
			await setDefaultDomain({ domain: row.domain })
			getTableData()
		},
	})
}

// Handle edit
const handleEdit = (row: MailDomain) => {
	formModalApi.setState({
		row,
		isEdit: true,
	})
	formModalApi.open()
}

// Handle delete
const handleDelete = (row: MailDomain) => {
	confirm({
		title: t('domain.delete.title'),
		content: t('domain.delete.confirm', { domain: row.domain }),
		confirmText: t('common.actions.delete'),
		confirmType: 'error',
		onConfirm: async () => {
			await deleteDomain({ domain: row.domain })
			getTableData()
		},
	})
}
</script>

<style lang="scss" scoped>
.domain-container {
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
