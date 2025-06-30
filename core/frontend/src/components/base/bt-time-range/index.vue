<template>
	<div class="bt-time-range">
		<n-radio-group v-model:value="activeFilter" @update:value="setFilter">
			<n-radio-button value="today">{{ $t('common.time.today') }}</n-radio-button>
			<n-radio-button value="yesterday">{{ $t('common.time.yesterday') }}</n-radio-button>
			<n-radio-button value="last7days">{{ $t('common.time.last7days') }}</n-radio-button>
		</n-radio-group>
		<n-date-picker
			:value="dateRange"
			type="daterange"
			:is-date-disabled="dateDisabled"
			@update:value="onDateRangeChange">
		</n-date-picker>
	</div>
</template>

<script lang="ts" setup>
import { addDays, startOfDay, endOfDay } from 'date-fns'
import { getDayTimeRange } from '@/utils'

const { defaultType } = defineProps({
	defaultType: {
		type: String,
		default: 'today',
	},
})

const dateRange = defineModel<[number, number]>('value')

const emit = defineEmits<{
	change: []
}>()

const activeFilter = ref(defaultType)

const getLastSevenDaysRange = (): [number, number] => {
	const today = new Date()
	const end = endOfDay(today)
	const start = startOfDay(addDays(today, -6))
	return [start.getTime(), end.getTime()]
}

const setFilter = (filter: string) => {
	switch (filter) {
		case 'today':
			dateRange.value = getDayTimeRange()
			break
		case 'yesterday':
			dateRange.value = getDayTimeRange(addDays(new Date(), -1))
			break
		case 'last7days':
			dateRange.value = getLastSevenDaysRange()
			break
	}
	emit('change')
}

const dateDisabled = (ts: number) => {
	const today = new Date()
	const before = addDays(today, -30)
	return startOfDay(before).getTime() > ts || endOfDay(today).getTime() < ts
}

const onDateRangeChange = (val: [number, number]) => {
	activeFilter.value = 'custom'
	dateRange.value = [startOfDay(val[0]).getTime(), endOfDay(val[1]).getTime()]
	emit('change')
}

setFilter(defaultType)
</script>

<style lang="scss" scoped>
.bt-time-range {
	display: flex;
	gap: 8px;
}
</style>
