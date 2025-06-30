<template>
	<div>
		<n-data-table :loading="loading" :data="serviceList" :columns="columns"></n-data-table>
		<config-modal></config-modal>
	</div>
</template>

<script lang="tsx" setup>
import { NButton, NTag, DataTableColumns, NFlex } from 'naive-ui'
import { confirm, isArray } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { getServiceList, restartService } from '@/api/modules/settings'
import type { DockerService } from '../types/common'

import ServiceConfig from './ServiceConfig.vue'

const { t } = useI18n()

const router = useRouter()

const loading = ref(false)
const serviceList = ref<DockerService[]>([])

const getConfigDisabled = (name: string) => {
	return name.includes('rspamd') || name.includes('postfix') || name.includes('dovecot')
}

const columns = ref<DataTableColumns<DockerService>>([
	{
		key: 'Names',
		title: t('settings.service.columns.name'),
		render: row => {
			return row.Names[0]
		},
	},
	{
		key: 'State',
		title: t('settings.service.columns.status'),
		render: row => {
			return (
				<NTag size="small" bordered={false} type={row.State === 'running' ? 'success' : 'error'}>
					{row.State === 'running'
						? t('settings.service.status.running')
						: t('settings.service.status.stopped')}
				</NTag>
			)
		},
	},
	{
		key: 'action',
		title: t('common.columns.actions'),
		align: 'right',
		render: row => (
			<NFlex inline={true}>
				{row.Names[0].includes('rspamd') ? (
					<NButton
						type="primary"
						text
						onClick={() => {
							router.push('/settings/rspamd')
						}}>
						GUI
					</NButton>
				) : null}
				<NButton
					type="primary"
					text
					disabled={!getConfigDisabled(row.Names[0])}
					onClick={() => {
						handleShowConfig(row)
					}}>
					{t('common.actions.config')}
				</NButton>
				<NButton
					type="primary"
					text
					onClick={() => {
						handleRestart(row)
					}}>
					{t('common.actions.restart')}
				</NButton>
			</NFlex>
		),
	},
])

const handleRestart = async (row: DockerService) => {
	confirm({
		title: t('settings.service.restart.title'),
		content: t('settings.service.restart.confirm', { name: row.Names[0] }),
		confirmText: t('common.actions.confirm'),
		onConfirm: async () => {
			await restartService({ container_id: row.Id })
			getList()
		},
	})
}

const [ConfigModal, configModalApi] = useModal({
	component: ServiceConfig,
})

const handleShowConfig = (row: DockerService) => {
	configModalApi.setState({ row: row })
	configModalApi.open()
}

const getList = async () => {
	try {
		loading.value = true
		const res = await getServiceList()
		if (isArray<DockerService>(res)) {
			serviceList.value = res
		}
	} finally {
		loading.value = false
	}
}

getList()
</script>

<style lang="scss" scoped>
.settings-title {
	margin-bottom: 16px;
	font-size: 16px;
	font-weight: 600;
	color: var(--color-text-2);
}
</style>
