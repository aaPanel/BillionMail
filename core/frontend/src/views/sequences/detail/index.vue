<template>
	<div class="p-24px">
		<n-page-header :title="sequence?.name || 'Sequence'" @back="router.back()">
			<template #extra>
				<NTag :type="sequence?.status === 'active' ? 'success' : 'default'" size="small">
					{{ sequence?.status }}
				</NTag>
			</template>
		</n-page-header>

		<n-card class="mt-16px" :title="t('sequences.detail.steps')">
			<n-data-table :columns="stepColumns" :data="steps" :loading="loading" />
		</n-card>

		<n-card class="mt-16px" :title="t('sequences.detail.stats')">
			<n-grid :cols="4" :x-gap="16">
				<n-gi>
					<n-statistic :label="t('sequences.columns.sent')" :value="sequence?.total_sent || 0" />
				</n-gi>
				<n-gi>
					<n-statistic :label="t('sequences.columns.opens')" :value="sequence?.total_opens || 0" />
				</n-gi>
				<n-gi>
					<n-statistic :label="t('sequences.columns.replies')" :value="sequence?.total_replies || 0" />
				</n-gi>
				<n-gi>
					<n-statistic label="Booked" :value="sequence?.total_booked || 0" />
				</n-gi>
			</n-grid>
		</n-card>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NTag } from 'naive-ui'
import { getSequenceDetail } from '@/api/modules/sequences'
import { formatTime } from '@/utils'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const loading = ref(true)
const sequence = ref<any>(null)
const steps = ref<any[]>([])

const stepColumns: DataTableColumns<any> = [
	{ title: t('sequences.detail.stepNumber'), key: 'step_number', width: 80 },
	{ title: t('sequences.detail.subject'), key: 'subject', ellipsis: { tooltip: true } },
	{
		title: t('sequences.detail.status'),
		key: 'status',
		width: 100,
		render: (row) => {
			const typeMap: Record<string, string> = { sent: 'success', pending: 'warning', skipped: 'default' }
			return <NTag type={typeMap[row.status] || 'default'} size="small">{row.status}</NTag>
		},
	},
	{
		title: t('sequences.detail.sentAt'),
		key: 'sent_at',
		width: 160,
		render: (row) => (row.sent_at ? formatTime(row.sent_at) : '-'),
	},
	{ title: t('sequences.detail.opened'), key: 'opened', width: 80, render: (row) => (row.opened ? 'Yes' : '-') },
	{ title: t('sequences.detail.clicked'), key: 'clicked', width: 80, render: (row) => (row.clicked ? 'Yes' : '-') },
	{ title: t('sequences.detail.replied'), key: 'replied', width: 80, render: (row) => (row.replied ? 'Yes' : '-') },
]

onMounted(async () => {
	try {
		const id = route.params.id as string
		const data = await getSequenceDetail(id)
		sequence.value = data.sequence
		steps.value = data.steps || []
	} finally {
		loading.value = false
	}
})
</script>
