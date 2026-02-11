<template>
	<div class="p-24px">
		<n-grid :cols="2" :x-gap="16">
			<!-- Scraper Config -->
			<n-gi>
				<n-card :title="t('enrichment.scraper.title')">
					<n-form :model="scraperForm" label-placement="left" label-width="120px">
						<n-form-item :label="t('enrichment.scraper.query')">
							<n-input v-model:value="scraperForm.query" :placeholder="t('enrichment.scraper.queryPlaceholder')" />
						</n-form-item>
						<n-form-item :label="t('enrichment.scraper.location')">
							<n-input v-model:value="scraperForm.location" placeholder="Austin, TX" />
						</n-form-item>
						<n-form-item :label="t('enrichment.scraper.maxResults')">
							<n-input-number v-model:value="scraperForm.max_results" :min="1" :max="500" />
						</n-form-item>
						<n-form-item>
							<n-button type="primary" :loading="scraping" @click="handleStartScrape">
								{{ t('enrichment.scraper.start') }}
							</n-button>
						</n-form-item>
					</n-form>
				</n-card>
			</n-gi>

			<!-- Pipeline Status -->
			<n-gi>
				<n-card :title="t('enrichment.pipeline.title')">
					<n-descriptions :column="1" bordered>
						<n-descriptions-item :label="t('enrichment.pipeline.scraped')">
							<n-tag>{{ pipelineStats.scraped || 0 }}</n-tag>
						</n-descriptions-item>
						<n-descriptions-item :label="t('enrichment.pipeline.enriching')">
							<n-tag type="warning">{{ pipelineStats.enriching || 0 }}</n-tag>
						</n-descriptions-item>
						<n-descriptions-item :label="t('enrichment.pipeline.personalized')">
							<n-tag type="success">{{ pipelineStats.personalized || 0 }}</n-tag>
						</n-descriptions-item>
						<n-descriptions-item :label="t('enrichment.pipeline.queued')">
							<n-tag type="info">{{ pipelineStats.queued || 0 }}</n-tag>
						</n-descriptions-item>
					</n-descriptions>
					<n-button class="mt-12px" @click="refreshPipeline">
						{{ t('common.actions.refresh') }}
					</n-button>
				</n-card>
			</n-gi>
		</n-grid>
	</div>
</template>

<script lang="ts" setup>
import { getScraperStatus, startScrapeJob } from '@/api/modules/enrichment'

const { t } = useI18n()

const scraperForm = reactive({
	query: '',
	location: '',
	max_results: 50,
})

const scraping = ref(false)
const pipelineStats = ref<Record<string, number>>({})

const handleStartScrape = async () => {
	scraping.value = true
	try {
		await startScrapeJob(scraperForm)
	} finally {
		scraping.value = false
	}
}

const refreshPipeline = async () => {
	const data = await getScraperStatus()
	pipelineStats.value = data || {}
}

onMounted(refreshPipeline)
</script>
