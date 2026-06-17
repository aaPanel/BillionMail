<template>
	<n-popover
		v-if="count !== null"
		trigger="manual"
		:show="showCount"
		placement="top"
		:show-arrow="true">
		<template #trigger>
			<n-card
				class="metric-card is-hoverable"
				:bordered="false"
				tabindex="0"
				role="button"
				:aria-label="`${title}: ${countLabel}`"
				@mouseenter="showCount = true"
				@mouseleave="showCount = false"
				@focus="showCount = true"
				@blur="showCount = false"
				@click="toggleCount">
				<div class="title">{{ title }}</div>
				<div class="value" :style="{ color: textColor }">{{ value }}{{ unit }}</div>
			</n-card>
		</template>
		<div class="count-popover">{{ title }}: {{ countLabel }}</div>
	</n-popover>

	<n-card v-else class="metric-card" :bordered="false">
		<div class="title">{{ title }}</div>
		<div class="value" :style="{ color: textColor }">{{ value }}{{ unit }}</div>
	</n-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const { t } = useI18n()

const props = defineProps({
	title: {
		type: String,
		default: '',
	},
	value: {
		type: Number,
		default: 0,
	},
	unit: {
		type: String,
		default: '',
	},
	count: {
		type: Number,
		default: null,
	},
	textColor: {
		type: String,
	},
})

const showCount = ref(false)

const countLabel = computed(() => t('overview.metricCount', { count: props.count ?? 0 }))

const toggleCount = () => {
	showCount.value = !showCount.value
}
</script>

<style lang="scss" scoped>
.metric-card {
	--n-padding-top: 24px;
	--n-padding-bottom: 24px;
	--n-text-color: var(--color-card-text-1);

	&.is-hoverable {
		cursor: pointer;
	}
}

.title {
	margin-bottom: 15px;
	font-size: 18px;
	font-weight: 400;
}

.value {
	font-size: 20px;
	line-height: 1.2;
}

.count-popover {
	font-size: 13px;
}
</style>
