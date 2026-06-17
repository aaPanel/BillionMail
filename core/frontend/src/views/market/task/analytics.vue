<template>
	<div class="p-20px">
		<div class="flex justify-between items-center mb-20px">
			<n-breadcrumb>
				<n-breadcrumb-item>
					<router-link to="/market/task">{{ $t('market.task.title') }}</router-link>
				</n-breadcrumb-item>
				<n-breadcrumb-item>{{ $t('market.task.actions.analytics') }}</n-breadcrumb-item>
				<n-breadcrumb-item>
					<n-ellipsis style="max-width: 300px">
						<span class="font-bold text-basic">{{ subject || '--' }}</span>
					</n-ellipsis>
				</n-breadcrumb-item>
			</n-breadcrumb>
			<bt-time-range v-model:value="dateRange" default-type="last7days" @change="fetchOverviewData">
			</bt-time-range>
		</div>

		<!-- Main Metric Cards -->
		<div class="metrics-cards">
			<metric-card
				v-for="(item, key) in rateData"
				:key="key"
				:title="item.label"
				:value="item.value"
				:count="item.count"
				:unit="item.unit">
			</metric-card>
		</div>

		<!-- Data Details Area -->
		<div class="detail-row">
			<!-- Left: Mail Provider Table -->
			<n-card class="provider-table-card" :title="$t('overview.mailProviders')">
				<provider-table v-model:value="providers" />
			</n-card>

			<!-- Right: Today's Sending Statistics -->
			<n-card class="send-today-card" :title="$t('overview.sendStats')">
				<send-today-stats :data="sendMail" />
			</n-card>
		</div>

		<!-- Bottom Rate Charts Area -->
		<div class="rate-charts-card">
			<rate-chart-panel :bounce="bounceRate" :click="clickRate" :open="openRate" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { getDayTimeRange, getNumber, isArray, isObject } from '@/utils'
import { getTaskDetails, getTaskOverview } from '@/api/modules/market/task'
import type { MailOverview, MailProvider, RateData } from '@/views/overview/types'

import MetricCard from '@/views/overview/components/MetricCard.vue'
import ProviderTable from '@/views/overview/components/ProviderTable.vue'
import SendTodayStats from '@/views/overview/components/SendTodayStats.vue'
import RateChartPanel from '@/views/overview/components/RateChartPanel.vue'

const route = useRoute()

const { t } = useI18n()

const id = computed(() => {
	return getNumber(route.params.id || '0')
})

const dateRange = ref(getDayTimeRange())

const providers = ref<MailProvider[]>([])

const rateData = reactive<RateData>({
	delivery_rate: { label: t('overview.delivered'), value: 0, unit: '%', count: 0 },
	open_rate: { label: t('overview.opened'), value: 0, unit: '%', count: 0 },
	click_rate: { label: t('overview.clicked'), value: 0, unit: '%', count: 0 },
	bounce_rate: { label: t('overview.bounced'), value: 0, unit: '%', count: 0 },
})

const sendMail = ref<MailOverview['send_mail_chart']>({
	column_type: 'hourly',
	dashboard: {
		delivered: 0,
		delivery_rate: 0,
		failed: 0,
		failure_rate: 0,
		sends: 0,
	},
	data: [],
})

const bounceRate = ref<MailOverview['bounce_rate_chart']>({
	column_type: 'hourly',
	data: [],
})

const clickRate = ref<MailOverview['click_rate_chart']>({
	column_type: 'hourly',
	data: [],
})

const openRate = ref<MailOverview['open_rate_chart']>({
	column_type: 'hourly',
	data: [],
})

// Function to update rate data
const updateRateData = (dashboard: MailOverview['dashboard']) => {
	Object.entries(dashboard).forEach(([key, value]) => {
		if (key in rateData) {
			rateData[key].value = value
		}
	})

	rateData.delivery_rate.count = dashboard.delivered
	rateData.open_rate.count = dashboard.opened
	rateData.click_rate.count = dashboard.clicked
	rateData.bounce_rate.count = dashboard.bounced
}

async function fetchOverviewData() {
	const res = await getTaskOverview({
		task_id: id.value,
		start_time: Math.floor(dateRange.value[0] / 1000),
		end_time: Math.floor(dateRange.value[1] / 1000),
	})

	if (isObject<MailOverview>(res)) {
		updateRateData(res.dashboard)

		providers.value = isArray(res.mail_providers) ? res.mail_providers : []
		sendMail.value = res.send_mail_chart
		bounceRate.value = res.bounce_rate_chart
		clickRate.value = res.click_rate_chart
		openRate.value = res.open_rate_chart
	}
}

const subject = ref('')

const getDetails = async () => {
	const res = await getTaskDetails({ id: id.value })
	if (isObject<{ subject: string }>(res)) {
		subject.value = res.subject
	}
}

getDetails()
</script>

<style lang="scss" scoped>
.metrics-cards {
	display: flex;
	justify-content: space-between;
	gap: 20px;
	margin-bottom: 20px;
}

.detail-row {
	display: flex;
	gap: 20px;
	margin-bottom: 20px;

	.provider-table-card,
	.send-today-card {
		flex: 1;
	}
}

.rate-charts-card {
	display: flex;
	gap: 20px;
}
</style>
