<template>
	<div class="flex gap-16px mb-24px">
		<n-card class="flex-1">
			<div class="title">
				<i class="i-fa:paper-plane"></i>
				<span>{{ $t('api.overview.totalSend') }}</span>
			</div>
			<div class="value">{{ overview.total_send }}</div>
		</n-card>
		<n-card class="flex-1">
			<div class="title">
				<i class="i-fa:envelope-open"></i>
				<span>{{ $t('api.overview.avgOpenRate') }}</span>
			</div>
			<div class="value">{{ overview.avg_open_rate }}%</div>
		</n-card>
		<n-card class="flex-1">
			<div class="title">
				<i class="i-fa:mouse-pointer"></i>
				<span>{{ $t('api.overview.avgClickRate') }}</span>
			</div>
			<div class="value">{{ overview.avg_click_rate }}%</div>
		</n-card>
		<n-card class="flex-1">
			<div class="title">
				<i class="i-fa:ban"></i>
				<span>{{ $t('api.overview.avgBounceRate') }}</span>
			</div>
			<div class="value">{{ overview.avg_bounce_rate }}%</div>
		</n-card>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue'
import { getOverviewStats } from '@/api/modules/api'
import { OverviewStats } from '../types/base'
import { isObject } from '@/utils'

const { time } = defineProps({
	time: {
		type: Object as PropType<[number, number]>,
		default: () => [0, 0],
	},
})

const overview = ref<OverviewStats>({
	total_send: 0,
	avg_delivery_rate: 0,
	avg_open_rate: 0,
	avg_click_rate: 0,
	avg_bounce_rate: 0,
	avg_unsub_rate: 0,
	total_unsubscribe: 0,
})

const getStats = async () => {
	const res = await getOverviewStats({
		start_time: Math.floor(time[0] / 1000),
		end_time: Math.floor(time[1] / 1000),
	})
	if (isObject<OverviewStats>(res)) {
		overview.value = res
	}
}

defineExpose({
	getStats,
})

onMounted(() => {
	getStats()
})
</script>

<style lang="scss" scoped>
.n-card {
	--n-padding-top: 20px;
	--n-padding-left: 20px;
	--n-padding-right: 20px;
	--n-padding-bottom: 20px;
}

.title {
	display: flex;
	align-items: center;
	margin-bottom: 8px;
	gap: 8px;
	font-size: 14px;
	color: var(--color-text-3);
}

.value {
	font-size: 28px;
	font-weight: 600;
	margin-bottom: 4px;
	font-family: 'Montserrat', sans-serif;
	color: var(--color-text-1);
}
</style>
